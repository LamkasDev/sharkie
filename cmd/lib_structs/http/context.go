package http

type HttpContext struct {
	Id uint32
}

func NewHttpContext(id uint32) *HttpContext {
	return &HttpContext{
		Id: id,
	}
}
