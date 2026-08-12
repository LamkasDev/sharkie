package http

type HttpEpoll struct {
	Id uint32
}

func NewHttpEpoll(id uint32) *HttpEpoll {
	return &HttpEpoll{
		Id: id,
	}
}
