package domain

type User struct {
	ID       string
	Email    string
	PassHash []byte
	IsAdmin  bool
}

type UserWithTokens struct {
	User
	AccessToken  string
	RefreshToken string
}
