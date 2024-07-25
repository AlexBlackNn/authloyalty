// Package auth_service SERVICE layer interface
package auth_service

import (
	"context"
)

type AuthorizationInterface interface {
	Login(
		ctx context.Context,
		email string,
		password string,
	) (accessToken string, refreshToken string, err error)
	Register(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	Logout(
		ctx context.Context,
		token string,
	) (success bool, err error)
	IsAdmin(
		ctx context.Context,
		userID int,
	) (success bool, err error)
	Validate(
		ctx context.Context,
		token string,
	) (success bool, err error)
	Refresh(
		ctx context.Context,
		token string,
	) (accessToken string, refreshToken string, err error)
}
