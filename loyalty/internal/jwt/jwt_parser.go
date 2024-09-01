package jwt

import (
	"errors"
	"log"

	"github.com/golang-jwt/jwt/v5"
)

func Parse(tokenString string) (string, string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Fatal(err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Fatal("invalid claims format")
	}
	userName, ok := claims["email"].(string)
	if !ok {
		return "", "", errors.New("email not found")
	}
	userId, ok := claims["uid"].(string)
	if !ok {
		return "", "", errors.New("uid not found")
	}

	return userId, userName, nil
}
