package integragtion_tests

import (
	"testing"
	"time"

	ssov1 "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	common2 "github.com/AlexBlackNn/authloyalty/sso/tests/integragtion_tests/common"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginHappyPath(t *testing.T) {
	ctx, testCommon := common2.New(t)

	respLogin, err := testCommon.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    "user@test.com",
		Password: "test",
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
	assert.Equal(t, "user@test.com", claims["email"].(string))

	// checking token expiration time might be only approximate
	const deltaSeconds = 1
	assert.InDelta(t, loginTime.Add(testCommon.Cfg.AccessTokenTtl).Unix(), claims["exp"].(float64), deltaSeconds)

}

func TestLoginFailCases(t *testing.T) {
	ctx, testCommon := common2.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Login with Empty Email",
			email:       "",
			password:    common2.RandomFakePassword(),
			expectedErr: "email is required",
		},
		{
			name:        "Login with Both Empty Email and Password",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testCommon.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
