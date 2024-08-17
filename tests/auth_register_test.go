package tests

import (
	ssov1 "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	"github.com/AlexBlackNn/authloyalty/tests/common"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRegister_HappyPath(t *testing.T) {
	ctx, testSuite := common.New(t)

	email := gofakeit.Email()
	password := common.RandomFakePassword()

	respReg, err := testSuite.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err) // if err exists - stop test
	assert.NotEmpty(t, respReg.GetUserId())
}

func TestDuplicatedRegistration(t *testing.T) {
	ctx, st := common.New(t)

	email := gofakeit.Email()
	pass := common.RandomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	require.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestAuthRegisterFailCases(t *testing.T) {
	ctx, st := common.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Register with Empty Email",
			email:       "",
			password:    common.RandomFakePassword(),
			expectedErr: "email is required",
		},
		{
			name:        "Register with Both Empty",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)

		})
	}
}
