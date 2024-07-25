package tests

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ssov1 "sso/protos/proto/sso/gen"
	"sso/tests/suite"
	"testing"
	"time"
)

func TestLogin_Login_HappyPath(t *testing.T) {
	ctx, testSuite := suite.New(t)

	respLogin, err := testSuite.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    "admin@test.com",
		Password: "test",
	})
	require.NoError(t, err)
	loginTime := time.Now() // to check token expiration time

	token := respLogin.GetAccessToken()

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(testSuite.Cfg.ServiceSecret), nil
	})
	require.NoError(t, err)

	// check validation
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// check out token consists correct information
	assert.Equal(t, int64(44), int64(claims["uid"].(float64)))
	assert.Equal(t, "admin@test.com", claims["email"].(string))

	// checking token expiration time might be only approximate
	const deltaSeconds = 1
	assert.InDelta(t, loginTime.Add(testSuite.Cfg.AccessTokenTtl).Unix(), claims["exp"].(float64), deltaSeconds)

}

//func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
//	ctx, st := suite.New(t)
//
//	email := gofakeit.Email()
//	pass := randomFakePassword()
//
//	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
//		Email:    email,
//		Password: pass,
//	})
//	require.NoError(t, err)
//	require.NotEmpty(t, respReg.GetUserId())
//
//	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
//		Email:    email,
//		Password: pass,
//	})
//	require.Error(t, err)
//	assert.Empty(t, respReg.GetUserId())
//	assert.ErrorContains(t, err, "user already exists")
//}
//
//func TestRegister_FailCases(t *testing.T) {
//	ctx, st := suite.New(t)
//
//	tests := []struct {
//		name        string
//		email       string
//		password    string
//		expectedErr string
//	}{
//		{
//			name:        "Register with Empty Password",
//			email:       gofakeit.Email(),
//			password:    "",
//			expectedErr: "password is required",
//		},
//		{
//			name:        "Register with Empty Email",
//			email:       "",
//			password:    randomFakePassword(),
//			expectedErr: "email is required",
//		},
//		{
//			name:        "Register with Both Empty",
//			email:       "",
//			password:    "",
//			expectedErr: "email is required",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
//				Email:    tt.email,
//				Password: tt.password,
//			})
//			require.Error(t, err)
//			require.Contains(t, err.Error(), tt.expectedErr)
//
//		})
//	}
//}
//
//func TestLogin_FailCases(t *testing.T) {
//	ctx, st := suite.New(t)
//
//	tests := []struct {
//		name        string
//		email       string
//		password    string
//		appID       int32
//		expectedErr string
//	}{
//		{
//			name:        "Login with Empty Password",
//			email:       gofakeit.Email(),
//			password:    "",
//			expectedErr: "password is required",
//		},
//		{
//			name:        "Login with Empty Email",
//			email:       "",
//			password:    randomFakePassword(),
//			expectedErr: "email is required",
//		},
//		{
//			name:        "Login with Both Empty Email and Password",
//			email:       "",
//			password:    "",
//			expectedErr: "email is required",
//		},
//		//{
//		//	name:        "Login with Non-Matching Password",
//		//	email:       gofakeit.Email(),
//		//	password:    randomFakePassword(),
//		//	expectedErr: "invalid email or password",
//		//},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
//				Email:    gofakeit.Email(),
//				Password: randomFakePassword(),
//			})
//			require.NoError(t, err)
//
//			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
//				Email:    tt.email,
//				Password: tt.password,
//			})
//			require.Error(t, err)
//			require.Contains(t, err.Error(), tt.expectedErr)
//		})
//	}
//}
