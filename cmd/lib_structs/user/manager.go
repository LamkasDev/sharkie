// Package user contains structs to manage PS4 users.
package user

import "sync"

var GlobalUserManager *UserManager

// UserManager keeps track of available users.
type UserManager struct {
	Users              map[int32]*User
	UsersByPlayerIndex map[uint8]*User
	Lock               sync.Mutex
}

func NewUserManager() *UserManager {
	manager := &UserManager{
		Users:              map[int32]*User{},
		UsersByPlayerIndex: map[uint8]*User{},
		Lock:               sync.Mutex{},
	}
	manager.AddUser(NewDefaultUser())

	return manager
}

func (um *UserManager) AddUser(user *User) {
	um.Lock.Lock()
	defer um.Lock.Unlock()
	um.Users[user.UserId] = user
	um.UsersByPlayerIndex[user.PlayerIndex] = user
}

func (um *UserManager) GetUser(userId int32) *User {
	um.Lock.Lock()
	defer um.Lock.Unlock()
	return um.Users[userId]
}

func (um *UserManager) GetUserByPlayerIndex(playerIndex uint8) *User {
	um.Lock.Lock()
	defer um.Lock.Unlock()
	return um.UsersByPlayerIndex[playerIndex]
}

func (um *UserManager) GetInitialUser() *User {
	return um.GetUserByPlayerIndex(1)
}

func SetupUserManager() {
	GlobalUserManager = NewUserManager()
}
