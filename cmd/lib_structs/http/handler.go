// Package http contains structs to emulate http connections.
package http

import "sync"

var GlobalHttpHandler *HttpHandler

// HttpHandler keeps state of http contexts.
type HttpHandler struct {
	Contexts      map[uint32]*HttpContext
	NextContextId uint32

	Templates          map[uint32]*HttpTemplate
	Connections        map[uint32]*HttpConnection
	Requests           map[uint32]*HttpRequest
	Epolls             map[uint32]*HttpEpoll
	NextObjectId       uint32
	AcceptEncodingGzip bool

	Lock sync.RWMutex
}

func NewHttpHandler() *HttpHandler {
	return &HttpHandler{
		Contexts:           map[uint32]*HttpContext{},
		NextContextId:      0x1001,
		Templates:          map[uint32]*HttpTemplate{},
		Connections:        map[uint32]*HttpConnection{},
		Requests:           map[uint32]*HttpRequest{},
		Epolls:             map[uint32]*HttpEpoll{},
		NextObjectId:       0x1001,
		AcceptEncodingGzip: true,
		Lock:               sync.RWMutex{},
	}
}

func (handler *HttpHandler) CreateContext() *HttpContext {
	handler.Lock.Lock()
	defer handler.Lock.Unlock()
	context := NewHttpContext(handler.NextObjectId)
	handler.Contexts[context.Id] = context
	handler.NextContextId++

	return context
}

func (handler *HttpHandler) GetContext(id uint32) *HttpContext {
	handler.Lock.RLock()
	defer handler.Lock.RUnlock()
	return handler.Contexts[id]
}

func (handler *HttpHandler) CreateTemplate() *HttpTemplate {
	handler.Lock.Lock()
	defer handler.Lock.Unlock()
	template := NewHttpTemplate(handler.NextObjectId)
	template.Settings.AcceptEncodingGzip = handler.AcceptEncodingGzip
	handler.Templates[template.Id] = template
	handler.NextObjectId++

	return template
}

func (handler *HttpHandler) GetTemplate(id uint32) *HttpTemplate {
	handler.Lock.RLock()
	defer handler.Lock.RUnlock()
	return handler.Templates[id]
}

func (handler *HttpHandler) CreateConnection() *HttpConnection {
	handler.Lock.Lock()
	defer handler.Lock.Unlock()
	connection := NewHttpConnection(handler.NextObjectId)
	handler.Connections[connection.Id] = connection
	handler.NextObjectId++

	return connection
}

func (handler *HttpHandler) GetConnection(id uint32) *HttpConnection {
	handler.Lock.RLock()
	defer handler.Lock.RUnlock()
	return handler.Connections[id]
}

func (handler *HttpHandler) CreateRequest() *HttpRequest {
	handler.Lock.Lock()
	defer handler.Lock.Unlock()
	request := NewHttpRequest(handler.NextObjectId)
	handler.Requests[request.Id] = request
	handler.NextObjectId++

	return request
}

func (handler *HttpHandler) GetRequest(id uint32) *HttpRequest {
	handler.Lock.RLock()
	defer handler.Lock.RUnlock()
	return handler.Requests[id]
}

func (handler *HttpHandler) DeleteRequest(id uint32) {
	handler.Lock.Lock()
	defer handler.Lock.Unlock()
	delete(handler.Requests, id)
}

func (handler *HttpHandler) CreateEpoll() *HttpEpoll {
	handler.Lock.Lock()
	defer handler.Lock.Unlock()
	epoll := NewHttpEpoll(handler.NextObjectId)
	handler.Epolls[epoll.Id] = epoll
	handler.NextObjectId++

	return epoll
}

func (handler *HttpHandler) GetEpoll(id uint32) *HttpEpoll {
	handler.Lock.RLock()
	defer handler.Lock.RUnlock()
	return handler.Epolls[id]
}

func (handler *HttpHandler) GetHeaderObject(id uint32) interface{} {
	handler.Lock.RLock()
	defer handler.Lock.RUnlock()
	template, ok := handler.Templates[id]
	if ok {
		return template
	}
	connection, ok := handler.Connections[id]
	if ok {
		return connection
	}
	request, ok := handler.Requests[id]
	if ok {
		return request
	}

	return nil
}

func SetupHttpHandler() {
	GlobalHttpHandler = NewHttpHandler()
}
