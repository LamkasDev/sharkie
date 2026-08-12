package http

type HttpConnection struct {
	Id         uint32
	TemplateId uint32

	Settings  *HttpSettings
	KeepAlive bool
	IsSecure  bool

	Url     string
	Scheme  string
	Host    string
	Port    uint16
	Headers map[string]string
}

func NewHttpConnection(id uint32) *HttpConnection {
	return &HttpConnection{
		Id:       id,
		Settings: NewHttpSettings(),
		Headers:  map[string]string{},
	}
}
