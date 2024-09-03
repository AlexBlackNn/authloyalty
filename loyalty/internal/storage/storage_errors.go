package storage

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrWrongParamType  = errors.New("wrong param type")
	ErrConnection      = errors.New("no connection")
	ErrNegativeBalance = errors.New("negative balance")
	ErrInternalErr     = errors.New("internal error")
)
