package jwt

import (
	"time"

	"github.com/AlexBlackNn/authloyalty/sso/internal/config"
	"github.com/AlexBlackNn/authloyalty/sso/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

// NewToken creates new JWT token for given user and app.
func NewToken(
	user domain.User,
	cfg *config.Config,
	tokenType string,
) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["token_type"] = tokenType
	claims["uid"] = user.ID
	claims["email"] = user.Email
	if tokenType == "access" {
		claims["exp"] = time.Now().Add(cfg.AccessTokenTtl).Unix()
	} else {
		claims["exp"] = time.Now().Add(cfg.RefreshTokenTtl).Unix()
	}
	tokenString, err := token.SignedString([]byte(cfg.ServiceSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
