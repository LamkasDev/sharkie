package http

type HttpSettings struct {
	ConnectionTimeout uint32 /* micros */
	SendTimeout       uint32 /* micros */
	ReceiveTimeout    uint32 /* micros */
	ResolveTimeout    uint32 /* micros */
	ResolveRetry      int32

	ReceiveBlockSize  uint64
	ResponseHeaderMax uint64

	AutoRedirect       bool
	InflateGzip        bool
	AcceptEncodingGzip bool
	NonBlocking        bool

	// SslFlags uint32
}

func NewHttpSettings() *HttpSettings {
	return &HttpSettings{
		AutoRedirect:       true,
		InflateGzip:        true,
		AcceptEncodingGzip: true,
	}
}
