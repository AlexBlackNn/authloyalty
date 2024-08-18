package tests

import (
	ssov1 "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	"github.com/AlexBlackNn/authloyalty/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsAdmin_HappyPath(t *testing.T) {
	ctx, testSuite := common.New(t)
	respIsAdmin, err := testSuite.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: "1",
	})
	require.NoError(t, err)
	isAdmin := respIsAdmin.GetIsAdmin()
	assert.Equal(t, true, isAdmin)

	respIsAdmin, err = testSuite.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: "2",
	})
	require.NoError(t, err)
	isAdmin = respIsAdmin.GetIsAdmin()
	assert.Equal(t, false, isAdmin)

}
