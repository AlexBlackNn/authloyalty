package tests

import (
	ssov1 "github.com/AlexBlackNn/authloyalty/protos/proto/sso/gen"
	"github.com/AlexBlackNn/authloyalty/tests/suite"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRegister_HappyPath(t *testing.T) {
	ctx, testSuite := suite.New(t)

	email := gofakeit.Email()
	password := suite.RandomFakePassword()

	respReg, err := testSuite.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err) // if err exists - stop test
	assert.NotEmpty(t, respReg.GetUserId())
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
