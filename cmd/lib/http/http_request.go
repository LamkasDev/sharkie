package http

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/config"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/http"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000011730
// __int64 __fastcall sceHttpCreateRequestWithURL(unsigned int, unsigned int, __int64, __int64)
func libSceHttp_sceHttpCreateRequestWithURL(connectionId uint32, method HttpMethod, url Cstring, contentLength uint64) uintptr {
	connection := GlobalHttpHandler.GetConnection(connectionId)
	if connection == nil {
		logger.Printf("%-132s %s failed due to invalid connection id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpCreateRequestWithURL"),
		)
		return 0x80431100
	}
	if url == nil {
		logger.Printf("%-132s %s failed due to invalid url.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpCreateRequestWithURL"),
		)
		return 0x80433060
	}
	if method > 8 || ((0x177>>uint(method&0x1f))&1) == 0 {
		logger.Printf("%-132s %s failed due to invalid method.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpCreateRequestWithURL"),
		)
		return 0x8043106B
	}

	// Create request.
	request := GlobalHttpHandler.CreateRequest()
	request.ConnectionId = connectionId
	request.Method = method
	request.Url = GoString(url)
	request.ContentLength = contentLength
	request.Settings = connection.Settings

	logger.Printf("%-132s %s created http request %s (connectionId=%s, method=%s, url=%s, contentLength=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpCreateRequestWithURL"),
		color.Yellow.Sprintf("0x%X", request.Id),
		color.Yellow.Sprintf("0x%X", connectionId),
		color.Yellow.Sprintf("0x%X", method),
		color.Blue.Sprint(request.Url),
		color.Green.Sprint(contentLength),
	)
	return uintptr(request.Id)
}

// 0x0000000000011CC0
// __int64 __fastcall sceHttpDeleteRequest(unsigned int)
func libSceHttp_sceHttpDeleteRequest(requestId uint32) uintptr {
	request := GlobalHttpHandler.GetRequest(requestId)
	if request == nil {
		logger.Printf("%-132s %s failed due to invalid request id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpDeleteRequest"),
		)
		return 0x80431100
	}
	request.Lock.Lock()
	request.State = HttpRequestStateAborted
	request.Cond.Broadcast()
	request.Lock.Unlock()
	GlobalHttpHandler.DeleteRequest(requestId)

	logger.Printf("%-132s %s deleted http request %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpDeleteRequest"),
		color.Yellow.Sprintf("0x%X", requestId),
	)
	return 0
}

// 0x0000000000011FF0
// __int64 __fastcall sceHttpSendRequest(unsigned int, __int64, __int64)
func libSceHttp_sceHttpSendRequest(requestId uint32, postDataPtr uintptr, size uint64) uintptr {
	request := GlobalHttpHandler.GetRequest(requestId)
	if request == nil {
		logger.Printf("%-132s %s failed due to invalid request id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpSendRequest"),
		)
		return 0x80431100
	}
	request.Lock.Lock()
	if request.State == HttpRequestStateSending || request.State == HttpRequestStateSent {
		logger.Printf("%-132s %s failed due to already send request.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpSendRequest"),
		)
		return 0x80431066
	}
	if request.State == HttpRequestStateAborted {
		logger.Printf("%-132s %s failed due to aborted request.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpSendRequest"),
		)
		return 0x80431080
	}
	request.State = HttpRequestStateSending

	// Construct a send plan.
	plan := HttpRequestSendPlan{
		Method:  request.Method,
		Path:    extractPathFromUrl(request.Url),
		Headers: map[string]string{},
	}

	// Fetch connection and template state.
	connection := GlobalHttpHandler.GetConnection(request.ConnectionId)
	if connection != nil {
		plan.Scheme = connection.Scheme
		plan.Host = connection.Host
		plan.Port = connection.Port

		// Inherit headers (template -> connection -> request).
		template := GlobalHttpHandler.GetTemplate(connection.TemplateId)
		if template != nil {
			maps.Copy(plan.Headers, template.Headers)
			// TODO: check ctx_has_loaded_certs.
		}
		maps.Copy(plan.Headers, connection.Headers)
	}
	maps.Copy(plan.Headers, request.Headers)

	// Extract Content-Type header.
	for k, v := range plan.Headers {
		if headerNameMatches(k, "Content-Type") {
			plan.ContentType = v
			break
		}
	}

	// Copy post data into a slice.
	if postDataPtr != 0 && size > 0 {
		plan.Body = make([]byte, size)
		copy(plan.Body, unsafe.Slice((*byte)(unsafe.Pointer(postDataPtr)), size))
	}
	request.Lock.Unlock()

	// Spawn async worker.
	callsiteText := emu.GlobalModuleManager.GetCallSiteText()
	go func(workerReqId uint32, workerPlan HttpRequestSendPlan) {
		var outResStatusCode HttpResponseStatusCode
		var outResHeadersBlob []byte
		var outResBody []byte
		var outResContentLength uint64
		outResContentLengthResult := uint64(0x80431071)
		workerErrno := uint64(0)

		// Create http.Request.
		if !config.GlobalConfig.NetworkEnabled {
			// Synthesize offline response.
			outResStatusCode = 0
			outResContentLength = 0
			outResContentLengthResult = uint64(0x80431071)
			outResBody = nil
			outResHeadersBlob = nil
			workerErrno = 0x80436002
		} else {
			methodString := workerPlan.Method.String()
			req, err := http.NewRequest(methodString, request.Url, bytes.NewReader(workerPlan.Body))
			if err == nil {
				for k, v := range workerPlan.Headers {
					req.Header.Set(k, v)
				}
				if workerPlan.ContentType != "" {
					req.Header.Set("Content-Type", workerPlan.ContentType)
				}

				// TODO: Handle redirects correctly (PS4 follows manually, drops body on 303, rewrites Host).
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					workerErrno = 0x80431050
				} else {
					defer resp.Body.Close()
					outResStatusCode = HttpResponseStatusCode(resp.StatusCode)

					if resp.ContentLength >= 0 {
						outResContentLength = uint64(resp.ContentLength)
						outResContentLengthResult = 0
					}

					var sb strings.Builder
					sb.WriteString(fmt.Sprintf("%s %d %s\r\n", resp.Proto, resp.StatusCode, http.StatusText(resp.StatusCode)))
					for k, vals := range resp.Header {
						for _, v := range vals {
							sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
						}
					}
					sb.WriteString("\r\n\x00")
					outResHeadersBlob = []byte(sb.String())

					// Read response body.
					if bodyBytes, err := io.ReadAll(resp.Body); err == nil {
						outResBody = bodyBytes
					}
				}
			} else {
				workerErrno = 0x804311FE
			}
		}

		// Refetch in case it was deleted during the request.
		currentRequest := GlobalHttpHandler.GetRequest(workerReqId)
		if currentRequest == nil {
			logger.Printf("%-132s %s finished sending deleted request.\n",
				callsiteText,
				color.Magenta.Sprint("sceHttpSendRequest"),
			)
			return
		}
		currentRequest.Lock.Lock()
		defer currentRequest.Lock.Unlock()
		if currentRequest.State == HttpRequestStateAborted {
			logger.Printf("%-132s %s finished sending aborted request.\n",
				callsiteText,
				color.Magenta.Sprint("sceHttpSendRequest"),
			)
			currentRequest.Cond.Broadcast() // Notify any waiting threads.
			return
		}

		// Update request state.
		currentRequest.State = HttpRequestStateSent
		currentRequest.LastErrno = workerErrno

		currentRequest.StatusCode = outResStatusCode
		currentRequest.AllHeadersBlob = outResHeadersBlob
		currentRequest.ResponseContentLength = outResContentLength
		currentRequest.ResponseContentLengthResult = outResContentLengthResult
		currentRequest.ResponseBody = outResBody
		currentRequest.ResponseBodyCursor = 0

		if workerErrno == 0 {
			logger.Printf("%-132s %s received reply for http request %s.\n",
				callsiteText,
				color.Magenta.Sprint("sceHttpSendRequest"),
				color.Yellow.Sprintf("0x%X", requestId),
			)
		} else {
			logger.Printf("%-132s %s failed receiving reply for http request %s.\n",
				callsiteText,
				color.Magenta.Sprint("sceHttpSendRequest"),
				color.Yellow.Sprintf("0x%X", requestId),
			)
		}

		// Notify sceHttpWaitRequest callers blocked on this request.
		currentRequest.Cond.Broadcast()
	}(requestId, plan)

	logger.Printf("%-132s %s sent http request %s (postDataPtr=%s, size=%s).\n",
		callsiteText,
		color.Magenta.Sprint("sceHttpSendRequest"),
		color.Yellow.Sprintf("0x%X", requestId),
		color.Yellow.Sprintf("0x%X", postDataPtr),
		color.Green.Sprint(size),
	)
	return 0
}

// 0x00000000000123F0
// __int64 __fastcall sceHttpGetStatusCode(unsigned int, _DWORD *)
func libSceHttp_sceHttpGetStatusCode(requestId uint32, statusCodePtr *HttpResponseStatusCode) uintptr {
	if statusCodePtr == nil {
		logger.Printf("%-132s %s failed due to invalid status code pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpGetStatusCode"),
		)
		return 0x804311FE
	}
	request := GlobalHttpHandler.GetRequest(requestId)
	if request == nil {
		logger.Printf("%-132s %s failed due to invalid request id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpGetStatusCode"),
		)
		return 0x80431100
	}
	err := waitForResponseReady(request)
	if err != 0 {
		return err
	}
	*statusCodePtr = request.StatusCode

	logger.Printf("%-132s %s returned http request %s's status code %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpGetStatusCode"),
		color.Yellow.Sprintf("0x%X", requestId),
		color.Green.Sprint(*statusCodePtr),
	)
	return 0
}

// 0x00000000000129C0
// __int64 __fastcall sceHttpGetAllResponseHeaders(unsigned int, _QWORD *, _QWORD *)
func libSceHttp_sceHttpGetAllResponseHeaders(requestId uint32, headersPtr *Cstring, headerSizePtr *uint64) uintptr {
	if headersPtr == nil || headerSizePtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpGetAllResponseHeaders"),
		)
		return 0x804311FE
	}
	request := GlobalHttpHandler.GetRequest(requestId)
	if request == nil {
		logger.Printf("%-132s %s failed due to invalid request id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpGetAllResponseHeaders"),
		)
		return 0x80431100
	}
	err := waitForResponseReady(request)
	if err != 0 {
		return err
	}
	if len(request.AllHeadersBlob) > 0 {
		*headersPtr = Cstring(unsafe.Pointer(&request.AllHeadersBlob[0]))
		*headerSizePtr = uint64(len(request.AllHeadersBlob) - 1)
	} else {
		*headersPtr = nil
		*headerSizePtr = 0
	}

	logger.Printf("%-132s %s returned http request %s's response headers (headerSize=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpGetAllResponseHeaders"),
		color.Yellow.Sprintf("0x%X", requestId),
		color.Green.Sprint(*headerSizePtr),
	)
	return 0
}

// 0x0000000000012220
// __int64 __fastcall sceHttpGetResponseContentLength(unsigned int, _DWORD *, _QWORD *)
func libSceHttp_sceHttpGetResponseContentLength(requestId uint32, resultPtr, contentLengthPtr *uint64) uintptr {
	if resultPtr == nil || contentLengthPtr == nil {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpGetResponseContentLength"),
		)
		return 0x804311FE
	}
	request := GlobalHttpHandler.GetRequest(requestId)
	if request == nil {
		logger.Printf("%-132s %s failed due to invalid request id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpGetResponseContentLength"),
		)
		return 0x80431100
	}
	err := waitForResponseReady(request)
	if err != 0 {
		return err
	}
	*resultPtr = request.ResponseContentLengthResult
	*contentLengthPtr = request.ResponseContentLength

	logger.Printf("%-132s %s returned http request %s's response content length %s (result=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpGetResponseContentLength"),
		color.Yellow.Sprintf("0x%X", requestId),
		color.Green.Sprint(*contentLengthPtr),
		color.Green.Sprint(*resultPtr),
	)
	return 0
}

// 0x0000000000012100
// __int64 __fastcall sceHttpReadData(unsigned int, __int64, __int64)
func libSceHttp_sceHttpReadData(requestId uint32, dataPtr uintptr, size uint64) uintptr {
	if dataPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid data pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpReadData"),
		)
		return 0x804311FE
	}
	request := GlobalHttpHandler.GetRequest(requestId)
	if request == nil {
		logger.Printf("%-132s %s failed due to invalid request id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpGetResponseContentLength"),
		)
		return 0x80431100
	}
	err := waitForResponseReady(request)
	if err != 0 {
		return err
	}
	request.Lock.Lock()
	defer request.Lock.Unlock()

	// Copy response body, advance cursor.
	bodySize := uint64(len(request.ResponseBody))
	remaining := max(bodySize-request.ResponseBodyCursor, 0)
	toCopy := min(remaining, size)
	if toCopy > 0 {
		dataSlice := unsafe.Slice((*byte)(unsafe.Pointer(dataPtr)), toCopy)
		copy(dataSlice, request.ResponseBody[request.ResponseBodyCursor:request.ResponseBodyCursor+toCopy])
		request.ResponseBodyCursor += toCopy
	}

	logger.Printf("%-132s %s read %s bytes of http request %s's response body (dataPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpReadData"),
		color.Green.Sprint(toCopy),
		color.Yellow.Sprintf("0x%X", requestId),
		color.Yellow.Sprintf("0x%X", dataPtr),
	)
	return uintptr(toCopy)
}

func waitForResponseReady(req *HttpRequest) uintptr {
	req.Lock.Lock()
	defer req.Lock.Unlock()
	if req.State == HttpRequestStateCreated {
		logger.Printf("%-132s %s failed due to unsent request.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("waitForResponseReady"),
		)
		return 0x8043105C
	}
	for req.State == HttpRequestStateSending {
		req.Cond.Wait()
	}
	if req.State == HttpRequestStateAborted {
		logger.Printf("%-132s %s failed due to aborted request.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("waitForResponseReady"),
		)
		return 0x80431080
	}
	if req.LastErrno != 0 && req.StatusCode == 0 {
		logger.Printf("%-132s %s failed due to transport failure.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("waitForResponseReady"),
		)
		return 0x8043105C // transport failure.
	}

	return 0
}

// extractPathFromUrl gets the path and query part from a full URL.
func extractPathFromUrl(u string) string {
	idx := strings.Index(u, "://")
	if idx == -1 {
		return "/"
	}
	authStart := idx + 3
	pathStart := strings.IndexByte(u[authStart:], '/')
	if pathStart == -1 {
		return "/"
	}

	return u[authStart+pathStart:]
}

// headerNameMatches does a case-insensitive match for HTTP headers.
func headerNameMatches(a, b string) bool {
	return strings.EqualFold(a, b)
}
