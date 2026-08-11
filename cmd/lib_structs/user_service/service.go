package user_service

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	fifo "github.com/foize/go.fifo"
)

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
	for _, user := range GlobalUserManager.GetLoggedInUsers() {
		GlobalUserService.EventQueue.Add(UserServiceEvent{Type: UserServiceEventTypeLogin, UserId: user.UserId})
	}
}
