package domain

type User struct {
	ID       string
	Email    string
	PassHash []byte
	IsAdmin  bool
	Name     string
	Birthday string
	Avatar   string
}

type UserWithTokens struct {
	User
	AccessToken  string
	RefreshToken string
}
