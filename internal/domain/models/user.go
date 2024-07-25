package models

type User struct {
	ID       int64
	Email    string
	PassHash []byte
	IsAdmin  bool
}

func (u *User) IsUserAmin() bool {
	return u.IsAdmin
}
