package http

import (
	"sync"
)

type HttpHeaderMode int32

const (
	HttpHeaderModeOverwrite = HttpHeaderMode(0)
	HttpHeaderModeAdd       = HttpHeaderMode(1)
)

type HttpMethod int32

const (
	HttpMethodGet     = HttpMethod(0)
	HttpMethodPost    = HttpMethod(1)
	HttpMethodHead    = HttpMethod(2)
	HttpMethodOptions = HttpMethod(3)
	HttpMethodPut     = HttpMethod(4)
	HttpMethodDelete  = HttpMethod(5)
	HttpMethodTrace   = HttpMethod(6)
	HttpMethodConnect = HttpMethod(7)
	HttpMethodCustom  = HttpMethod(8)
)

func (method HttpMethod) String() string {
	switch method {
	case HttpMethodGet:
		return "GET"
	case HttpMethodPost:
		return "POST"
	case HttpMethodHead:
		return "HEAD"
	case HttpMethodOptions:
		return "OPTIONS"
	case HttpMethodPut:
		return "PUT"
	case HttpMethodDelete:
		return "DELETE"
	case HttpMethodTrace:
		return "TRACE"
	case HttpMethodConnect:
		return "CONNECT"
	case HttpMethodCustom:
		return "PATCH" // Default for custom per PS4 behavior.
	default:
		return "GET"
	}
}

type HttpResponseStatusCode int32

type HttpRequestState uint32

const (
	HttpRequestStateCreated = HttpRequestState(iota)
	HttpRequestStateSending
	HttpRequestStateSent
	HttpRequestStateAborted
)

type HttpRequest struct {
	Id           uint32
	ConnectionId uint32

	Lock          sync.Mutex
	Settings      *HttpSettings
	ContentLength uint64

	State     HttpRequestState
	LastErrno uint64

	// Request fields.
	Method  HttpMethod
	Url     string
	Headers map[string]string

	// Response fields.
	Cond                        *sync.Cond
	StatusCode                  HttpResponseStatusCode
	AllHeadersBlob              []byte
	ResponseContentLength       uint64
	ResponseContentLengthResult uint64
	ResponseBody                []byte
	ResponseBodyCursor          uint64
}

func NewHttpRequest(id uint32) *HttpRequest {
	req := &HttpRequest{
		Id:       id,
		Lock:     sync.Mutex{},
		Settings: NewHttpSettings(),
		Headers:  map[string]string{},
	}
	req.Cond = sync.NewCond(&req.Lock)
	return req
}

type HttpRequestSendPlan struct {
	Method HttpMethod
	Path   string
	Scheme string
	Host   string
	Port   uint16

	ContentType string
	Headers     map[string]string
	Body        []byte
}
