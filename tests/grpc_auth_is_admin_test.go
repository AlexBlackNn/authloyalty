package tests

import (
	ssov1 "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	"github.com/AlexBlackNn/authloyalty/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsAdminHappyPath(t *testing.T) {
	ctx, testCommon := common.New(t)
	respIsAdmin, err := testCommon.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: "22f23689-9b67-4ef9-a693-5ef2d18ee111",
	})
	require.NoError(t, err)
	isAdmin := respIsAdmin.GetIsAdmin()
	assert.Equal(t, true, isAdmin)

	respIsAdmin, err = testCommon.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: "7c2ab9ec-bddf-43ff-96a5-ff1e0785c909",
	})
	require.NoError(t, err)
	isAdmin = respIsAdmin.GetIsAdmin()
	assert.Equal(t, false, isAdmin)

}
