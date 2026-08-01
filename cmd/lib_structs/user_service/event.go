package user_service

import . "github.com/LamkasDev/sharkie/cmd/lib_structs/user"

type UserServiceEventType int32

const (
	UserServiceEventTypeInvalid = UserServiceEventType(-1)
	UserServiceEventTypeLogin   = UserServiceEventType(0)
	UserServiceEventTypeLogout  = UserServiceEventType(1)
)

type UserServiceEvent struct {
	Type   UserServiceEventType
	UserId UserId
}
