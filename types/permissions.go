package types

// UserPerm user permission enum
type UserPerm string

// User permission enums
const (
	Admin     UserPerm = "Admin"
	Moderator UserPerm = "Moderator"
)

// UserPerms contains all user permissions
var UserPerms = []UserPerm{
	Admin,
	Moderator,
}

func (e UserPerm) UserString() string {
	return string(e)
}
