package authservice

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrTokenParsing       = errors.New("fail to parse token")
	ErrTokenTTLExpired    = errors.New("token ttl expired")
	ErrTokenWrongType     = errors.New("token wrong type")
)
