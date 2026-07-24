package user

type UserId int32

const (
	UserIdInvalid = UserId(-1)
)

type UserColor uint32

const (
	UserColorBlue  = UserColor(0)
	UserColorRed   = UserColor(1)
	UserColorGreen = UserColor(2)
	UserColorPink  = UserColor(3)
)

type User struct {
	UserId      UserId
	UserName    string
	UserColor   UserColor
	PlayerIndex uint8
	LoggedIn    bool
}

func NewDefaultUser() *User {
	return &User{
		UserId:      1000,
		UserName:    "sharkie",
		UserColor:   UserColorRed,
		PlayerIndex: 1,
		LoggedIn:    true,
	}
}

const MaxLoginUsers = 4

type LoginUserIdList struct {
	UserIds [MaxLoginUsers]UserId
}
