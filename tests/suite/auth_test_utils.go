package suite

import "github.com/brianvoe/gofakeit/v6"

const (
	PasswordDefaultLen = 10
)

func RandomFakePassword() string {
	return gofakeit.Password(
		true,
		true,
		true,
		true,
		false,
		PasswordDefaultLen,
	)
}
