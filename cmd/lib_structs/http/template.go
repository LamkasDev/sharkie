package http

type HttpTemplate struct {
	Id        uint32
	ContextId uint32

	Settings      *HttpSettings
	AutoProxyConf bool

	HttpVersion uint32
	UserAgent   string
	Headers     map[string]string
}

func NewHttpTemplate(id uint32) *HttpTemplate {
	return &HttpTemplate{
		Id:       id,
		Settings: NewHttpSettings(),
		Headers:  map[string]string{},
	}
}
