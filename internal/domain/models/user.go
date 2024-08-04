package models

type User struct {
	ID       string
	Email    string
	PassHash []byte
	IsAdmin  bool
}

func (u *User) IsUserAmin() bool {
	return u.IsAdmin
}
