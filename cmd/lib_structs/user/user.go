package user

type User struct {
	UserId      int32
	UserName    string
	UserColor   uint32
	PlayerIndex uint8
}

func NewDefaultUser() *User {
	return &User{
		UserId:      1000,
		UserName:    "sharkie",
		UserColor:   1,
		PlayerIndex: 1,
	}
}
