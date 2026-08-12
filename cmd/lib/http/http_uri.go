package http

import (
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/http"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000021B40
// __int64 __fastcall sceHttpUriParse(_DWORD *_RDI, unsigned __int8 *, _BYTE *, _QWORD *, unsigned __int64, __m128 _XMM0)
func libSceHttp_sceHttpUriParse(result *HttpUriElement, srcUri Cstring, poolPtr uintptr, require *uint64, prepare uint64) uintptr {
	if srcUri == nil {
		logger.Printf("%-132s %s failed due to invalid source uri.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpUriParse"),
		)
		return 0x80433060
	}
	writeOutput := (result != nil) && (poolPtr != 0)
	if !writeOutput && require == nil {
		logger.Printf("%-132s %s failed due to invalid result parameters.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpUriParse"),
		)
		return 0x804311FE
	}

	// Prepare parameters.
	if writeOutput {
		*result = HttpUriElement{}
	}
	src := GoString(srcUri)
	poolUsed := uint64(0)
	inputOffset := 0

	// Parse scheme.
	hasScheme := false
	schemeLength := 0
	{
		i := scanUntil(src, 0x20, isSchemeChar)
		if i > 0 && i < len(src) && src[i] == ':' && isAlpha(src[0]) {
			hasScheme = true
			schemeLength = i
		}
	}

	var schemeString string
	if hasScheme {
		schemeString = src[:schemeLength]
		poolUsed = uint64(schemeLength + 1)
		if writeOutput {
			if prepare < poolUsed {
				logger.Printf("%-132s %s failed due to invalid size.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sceHttpUriParse"),
				)
				return 0x80431022
			}
			dest := unsafe.Slice((*byte)(unsafe.Pointer(poolPtr)), poolUsed)
			copy(dest, schemeString)
			dest[schemeLength] = 0
			result.Scheme = Cstring(unsafe.Pointer(poolPtr))
		}
		inputOffset = schemeLength + 1
	} else {
		poolUsed = 1
		if writeOutput {
			if prepare < 2 { // this seems wrong, but it's like this.
				logger.Printf("%-132s %s failed due to invalid size.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sceHttpUriParse"),
				)
				return 0x80431022
			}
			*(*byte)(unsafe.Pointer(poolPtr)) = 0
			result.Scheme = Cstring(unsafe.Pointer(poolPtr))
		}
		inputOffset = 0
	}

	// Check slashes.
	slashCount := 0
	for inputOffset+slashCount < len(src) && src[inputOffset+slashCount] == '/' {
		slashCount++
	}
	if slashCount >= 2 {
		inputOffset += 2
	} else if writeOutput {
		result.Opaque = true
	}

	// Parse authority (username/password).
	authString := ""
	if inputOffset < len(src) {
		authString = src[inputOffset:]
	}

	seenColon, seenAt := false, false
	colonPos, atPos := 0, 0
	for scanPos := 0; scanPos < len(authString); scanPos++ {
		c := authString[scanPos]
		if c == '@' {
			seenAt = true
			atPos = scanPos
			break
		}
		if !seenColon && c == ':' {
			seenColon = true
			colonPos = scanPos
			continue
		}
		if isHighBit(c) || (!isAlnum(c) && !isUserinfoPunct(c)) {
			break
		}
	}

	var userString, passString string
	authAdvance := 0
	if seenAt {
		if seenColon {
			userString = authString[:colonPos]
			passString = authString[colonPos+1 : atPos]
		} else {
			userString = authString[:atPos]
			passString = ""
		}
		authAdvance = atPos + 1
	}

	// Write back authority (username/password).
	var err uintptr
	var ptr Cstring
	ptr, err = writePoolString(poolPtr, &poolUsed, prepare, writeOutput, userString)
	if err != 0 {
		return err
	}
	if writeOutput {
		result.Username = ptr
	}
	ptr, err = writePoolString(poolPtr, &poolUsed, prepare, writeOutput, passString)
	if err != 0 {
		return err
	}
	if writeOutput {
		result.Password = ptr
	}
	inputOffset += authAdvance

	// Parse host.
	hostScanLength, storedHostLength := 0, 0
	var hostString string
	if inputOffset < len(src) {
		first := src[inputOffset]
		switch first {
		case '.':
			// empty host.
		case '[': // IPv6.
			hostScanLength = 1
			for {
				if hostScanLength == 0xFF {
					logger.Printf("%-132s %s failed due to invalid ipv6 host size.\n",
						emu.GlobalModuleManager.GetCallSiteText(),
						color.Magenta.Sprint("sceHttpUriParse"),
					)
					return 0x80433060
				}
				if inputOffset+hostScanLength >= len(src) {
					break
				}
				c := src[inputOffset+hostScanLength]
				if isHighBit(c) || c == ']' {
					break
				}
				if !isIPv6HostChar(c) {
					break
				}
				hostScanLength++
			}
			if inputOffset+hostScanLength >= len(src) || src[inputOffset+hostScanLength] != ']' {
				logger.Printf("%-132s %s failed due to invalid ipv6 host.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sceHttpUriParse"),
				)
				return 0x80433060
			}
			storedHostLength = hostScanLength - 1
			hostString = src[inputOffset+1 : inputOffset+1+storedHostLength]
			hostScanLength++ // consume ']'.
		default: // normal host.
			hostScanLength = scanUntil(src[inputOffset:], 0xFF, func(c byte) bool {
				return !isHighBit(c) && isHostChar(c)
			})
			if hostScanLength == 0xFF {
				logger.Printf("%-132s %s failed due to invalid ipv4 host size.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sceHttpUriParse"),
				)
				return 0x80433060
			}
			storedHostLength = hostScanLength
			hostString = src[inputOffset : inputOffset+storedHostLength]
		}
	}

	// Write back host.
	ptr, err = writePoolString(poolPtr, &poolUsed, prepare, writeOutput, hostString)
	if err != 0 {
		return err
	}
	if writeOutput {
		result.Hostname = ptr
	}
	inputOffset += hostScanLength

	// Parse port.
	hasExplicitPort := false
	portValue := uint16(0)
	if inputOffset < len(src) && src[inputOffset] == ':' {
		inputOffset++
		digitsLen := 0
		port32 := uint32(0)
		for inputOffset+digitsLen < len(src) && digitsLen < 5 && isDigit(src[inputOffset+digitsLen]) {
			port32 = port32*10 + uint32(src[inputOffset+digitsLen]-'0')
			digitsLen++
		}
		if port32 >= 0x10000 {
			logger.Printf("%-132s %s failed due to invalid port number.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceHttpUriParse"),
			)
			return 0x80433060
		}
		if inputOffset+digitsLen < len(src) {
			afterPort := src[inputOffset+digitsLen]
			if afterPort != '/' {
				logger.Printf("%-132s %s failed due to invalid port.\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("sceHttpUriParse"),
				)
				return 0x80433060
			}
		}
		if digitsLen > 0 {
			hasExplicitPort = true
			portValue = uint16(port32)
		}
		inputOffset += digitsLen
	}

	if writeOutput {
		if hasExplicitPort {
			result.Port = portValue
		} else if hasScheme {
			scheme := strings.ToLower(schemeString)
			switch scheme {
			case "https":
				result.Port = 443
			case "http":
				result.Port = 80
			}
		}
	}

	// Parse path.
	pathLength := scanUntil(src[inputOffset:], 0x3FFF, func(c byte) bool {
		return c != '?' && c != '#'
	})
	if pathLength >= 0x3FFF {
		logger.Printf("%-132s %s failed due to invalid path size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpUriParse"),
		)
		return 0x80433060
	}
	pathString := src[inputOffset : inputOffset+pathLength]

	// Write back path.
	ptr, err = writePoolString(poolPtr, &poolUsed, prepare, writeOutput, pathString)
	if err != 0 {
		return err
	}
	if writeOutput {
		tmpBuf := make([]byte, pathLength+1)
		copy(tmpBuf, pathString)
		tmpBuf[pathLength] = 0

		tmpPtr := Cstring(unsafe.Pointer(&tmpBuf[0]))
		libSceHttp_sceHttpUriSweepPath(ptr, tmpPtr, uint64(pathLength+1))
		result.Path = ptr
	}
	inputOffset += pathLength

	// Parse query.
	queryLength := 0
	if inputOffset < len(src) && src[inputOffset] == '?' {
		queryLength = 1 + scanUntil(src[inputOffset+1:], 0x3FFF-1, func(c byte) bool {
			return c != '#'
		})
		if queryLength >= 0x3FFF {
			logger.Printf("%-132s %s failed due to invalid query size.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceHttpUriParse"),
			)
			return 0x80433060
		}
	}

	var queryString string
	if inputOffset <= len(src) {
		queryString = src[inputOffset : inputOffset+queryLength]
	}

	// Write back query.
	ptr, err = writePoolString(poolPtr, &poolUsed, prepare, writeOutput, queryString)
	if err != 0 {
		return err
	}
	if writeOutput {
		result.Query = ptr
	}
	inputOffset += queryLength

	// Parse fragment.
	fragLength := 0
	if inputOffset < len(src) && src[inputOffset] == '#' {
		fragLength = 1 + scanUntil(src[inputOffset+1:], 0x3FFF-1, func(c byte) bool {
			return true // anything until end.
		})
		if fragLength >= 0x3FFF {
			logger.Printf("%-132s %s failed due to invalid fragment size.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceHttpUriParse"),
			)
			return 0x80433060
		}
	}

	var fragString string
	if inputOffset <= len(src) {
		fragString = src[inputOffset : inputOffset+fragLength]
	}

	// Write back fragment.
	ptr, err = writePoolString(poolPtr, &poolUsed, prepare, writeOutput, fragString)
	if err != 0 {
		return err
	}
	if writeOutput {
		result.Fragment = ptr
	}

	// Write back pool.
	if require != nil {
		*require = poolUsed
	}

	logger.Printf("%-132s %s parsed uri (scheme=%s, hostname=%s, path=%s, query=%s, port=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpUriParse"),
		color.Blue.Sprint(schemeString),
		color.Blue.Sprint(hostString),
		color.Blue.Sprint(pathString),
		color.Blue.Sprint(queryString),
		color.Green.Sprint(portValue),
	)
	return 0
}

// TODO: create helpers for this.
// 0x0000000000022BA0
// __int64 __fastcall sceHttpUriSweepPath(__int64, __int64, __int64)
func libSceHttp_sceHttpUriSweepPath(dst Cstring, src Cstring, srcSize uint64) uintptr {
	if srcSize == 0 {
		return 0
	}
	if dst == nil || src == nil {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpUriSweepPath"),
		)
		return 0x804311FE
	}
	dstSlice := unsafe.Slice((*byte)(dst), srcSize)
	srcSlice := unsafe.Slice((*byte)(src), srcSize)

	// Non-absolute path.
	if srcSlice[0] != '/' {
		copyLen := srcSize - 1
		copy(dstSlice, srcSlice[:copyLen])
		dstSlice[copyLen] = 0
		return 0
	}

	// Absolute path (dst[0]='/', dst[1]='\0').
	dstSlice[0] = '/'
	dstSlice[1] = 0
	if srcSize-1 <= 1 {
		return 0
	}

	srcPos := uint64(1)
	segmentEnd := uint64(0) // acts as the `segmentEnd` pointer offset relative to dst.
	for srcPos < srcSize-1 {
		if srcSlice[srcPos] == '.' {
			if srcPos+1 < srcSize && srcSlice[srcPos+1] == '/' {
				// "./" - skip.
				srcPos += 2
				continue
			}
			if srcPos+2 < srcSize && srcSlice[srcPos+1] == '.' && srcSlice[srcPos+2] == '/' {
				// "../" - backup one segment.
				newSegmentEnd := uint64(0)
				if segmentEnd != 0 {
					dstSlice[segmentEnd] = 0 // Temporarily terminate for our backwards scan.

					// Find previous slash.
					prevSlash := -1
					for i := int(segmentEnd - 1); i >= 0; i-- {
						if dstSlice[i] == '/' {
							prevSlash = i
							break
						}
					}

					if prevSlash == -1 {
						newSegmentEnd = 0
					} else {
						dstSlice[prevSlash+1] = 0
						newSegmentEnd = uint64(prevSlash)
					}
				}
				srcPos += 3
				segmentEnd = newSegmentEnd
				continue
			}
		}

		// Find next slash.
		nextSlash := -1
		for i := srcPos; i < srcSize-1; i++ {
			if srcSlice[i] == '/' {
				nextSlash = int(i)
				break
			}
		}

		remaining := srcSize - srcPos - 1
		var copyLength uint64
		if nextSlash == -1 {
			copyLength = remaining
		} else {
			segLength := uint64(nextSlash) + 1 - srcPos
			if segLength <= remaining {
				copyLength = segLength
			} else {
				copyLength = remaining
			}
		}

		// Perform the segmented copy.
		copy(dstSlice[segmentEnd+1:], srcSlice[srcPos:srcPos+copyLength])
		dstSlice[segmentEnd+copyLength+1] = 0

		segmentEnd += copyLength
		srcPos += copyLength
	}

	return 0
}

// scanUntil returns the number of leading bytes that satisfy pred, capped at max.
// Stops at the first byte that fails the predicate or at max.
func scanUntil(s string, max int, pred func(byte) bool) int {
	n := 0
	for n < len(s) && n < max && pred(s[n]) {
		n++
	}
	return n
}

// isSchemeChar matches the original scheme character class.
func isSchemeChar(c byte) bool {
	return isAlnum(c) || c == '+' || c == '-' || c == '.'
}

// isHostChar matches the normal (non-IPv6) host character class.
func isHostChar(c byte) bool {
	return isAlnum(c) || c == '-' || c == '.' || c == '_'
}

// isIPv6HostChar matches the IPv6 bracket contents character class.
func isIPv6HostChar(c byte) bool {
	return isAlnum(c) || c == '-' || c == '.' || c == '_' || c == ':'
}

func isAlpha(c byte) bool   { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isDigit(c byte) bool   { return c >= '0' && c <= '9' }
func isAlnum(c byte) bool   { return isAlpha(c) || isDigit(c) }
func isHighBit(c byte) bool { return c > 127 }
func isUserinfoPunct(c byte) bool {
	switch c {
	case 0x21, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x3a, 0x3b, 0x3d, 0x5f, 0x7e:
		return true
	}
	return false
}

// writePoolString writes a Go string into the pre-allocated PS4 memory pool; appending a NULL terminator.
func writePoolString(poolPtr uintptr, poolUsed *uint64, prepare uint64, writeOutput bool, str string) (Cstring, uintptr) {
	needed := uint64(len(str) + 1)
	var ptr Cstring
	if writeOutput {
		if prepare-*poolUsed < needed {
			logger.Printf("%-132s %s failed due to invalid size.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("writePoolString"),
			)
			return nil, 0x80431022
		}
		dest := unsafe.Slice((*byte)(unsafe.Pointer(poolPtr+uintptr(*poolUsed))), needed)
		copy(dest, str)
		dest[len(str)] = 0
		ptr = Cstring(unsafe.Pointer(poolPtr + uintptr(*poolUsed)))
	}

	*poolUsed += needed
	return ptr, 0
}
