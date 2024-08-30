package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
)

func JWTParse(tokenString string) (string, string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Fatal(err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Fatal("invalid claims format")
	}

	for key, value := range claims {
		fmt.Printf("%s = %v\n", key, value)
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
