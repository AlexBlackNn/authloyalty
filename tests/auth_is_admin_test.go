package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ssov1 "sso/protos/proto/sso/gen"
	"sso/tests/suite"
	"testing"
)

func TestIsAdmin_HappyPath(t *testing.T) {
	ctx, testSuite := suite.New(t)
	respIsAdmin, err := testSuite.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: 1,
	})
	require.NoError(t, err)
	isAdmin := respIsAdmin.GetIsAdmin()
	assert.Equal(t, true, isAdmin)

	respIsAdmin, err = testSuite.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: 2,
	})
	require.NoError(t, err)
	isAdmin = respIsAdmin.GetIsAdmin()
	assert.Equal(t, false, isAdmin)

}
