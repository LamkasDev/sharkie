package user_service

import fifo "github.com/foize/go.fifo"

var GlobalUserService *UserService

type UserService struct {
	EventQueue *fifo.Queue
}

func NewUserService() *UserService {
	return &UserService{
		EventQueue: fifo.NewQueue(),
	}
}

func SetupUserService() {
	GlobalUserService = NewUserService()
}
