// Package user contains structs to manage PS4 users.
package user

import (
	"sync"
)

var GlobalUserManager *UserManager

// UserManager keeps track of available users.
type UserManager struct {
	Users              map[UserId]*User
	UsersByPlayerIndex map[uint8]*User
	Lock               sync.Mutex
}

func NewUserManager() *UserManager {
	manager := &UserManager{
		Users:              map[UserId]*User{},
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

func (um *UserManager) GetUser(userId UserId) *User {
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

func (um *UserManager) GetLoggedInUsers() []*User {
	um.Lock.Lock()
	defer um.Lock.Unlock()
	var loggedInUsers []*User
	for _, user := range um.Users {
		if user.LoggedIn {
			loggedInUsers = append(loggedInUsers, user)
		}
	}

	return loggedInUsers
}

func SetupUserManager() {
	GlobalUserManager = NewUserManager()
}
