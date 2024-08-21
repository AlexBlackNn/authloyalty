package tests

import (
	ssov1 "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	"github.com/AlexBlackNn/authloyalty/tests/common"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRegisterLoginHappyPath(t *testing.T) {
	ctx, testCommon := common.New(t)

	email := gofakeit.Email()
	password := common.RandomFakePassword()
	respReg, err := testCommon.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err) // if err exists - stop test
	assert.NotEmpty(t, respReg.GetUserId())
	respLogin, err := testCommon.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	loginTime := time.Now() // to check token expiration time

	token := respLogin.GetAccessToken()
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(testCommon.Cfg.ServiceSecret), nil
	})
	require.NoError(t, err)

	// check validation
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// check out token consists correct information
	assert.Equal(t, respReg.GetUserId(), claims["uid"].(string))
	assert.Equal(t, email, claims["email"].(string))

	// checking token expiration time might be only approximate
	const deltaSeconds = 1
	assert.InDelta(t, loginTime.Add(testCommon.Cfg.AccessTokenTtl).Unix(), claims["exp"].(float64), deltaSeconds)

}
