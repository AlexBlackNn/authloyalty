package storage

import "errors"

var (
	ErrUserExists     = errors.New("user already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrAppNotFound    = errors.New("app not found")
	ErrWrongParamType = errors.New("wrong param type")
	ErrConnection     = errors.New("no connection")
)
