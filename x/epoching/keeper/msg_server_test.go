package keeper_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// TODO (fuzz tests): replace the following tests with fuzz ones
func TestMsgWrappedDelegate(t *testing.T) {
	_, ctx, _, msgSrvr, queryClient, _ := setupTestKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	testCases := []struct {
		name      string
		req       *stakingtypes.MsgDelegate
		expectErr bool
	}{
		{
			"empty wrapped msg",
			&stakingtypes.MsgDelegate{},
			false,
		},
	}
	for _, tc := range testCases {

		wrappedMsg := types.NewMsgWrappedDelegate(tc.req)
		_, err := msgSrvr.WrappedDelegate(wctx, wrappedMsg)
		require.NoError(t, err)

		resp, err := queryClient.EpochMsgs(wctx, &types.QueryEpochMsgsRequest{
			Pagination: &query.PageRequest{},
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Msgs))

		if tc.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestMsgWrappedUndelegate(t *testing.T) {
	_, ctx, _, msgSrvr, queryClient, _ := setupTestKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	testCases := []struct {
		name      string
		req       *stakingtypes.MsgUndelegate
		expectErr bool
	}{
		{
			"empty wrapped msg",
			&stakingtypes.MsgUndelegate{},
			false,
		},
	}
	for _, tc := range testCases {
		wrappedMsg := types.NewMsgWrappedUndelegate(tc.req)
		_, err := msgSrvr.WrappedUndelegate(wctx, wrappedMsg)
		require.NoError(t, err)

		resp, err := queryClient.EpochMsgs(wctx, &types.QueryEpochMsgsRequest{
			Pagination: &query.PageRequest{},
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Msgs))

		if tc.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestMsgWrappedBeginRedelegate(t *testing.T) {
	_, ctx, _, msgSrvr, queryClient, _ := setupTestKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	testCases := []struct {
		name      string
		req       *stakingtypes.MsgBeginRedelegate
		expectErr bool
	}{
		{
			"empty wrapped msg",
			&stakingtypes.MsgBeginRedelegate{},
			false,
		},
	}
	for _, tc := range testCases {
		wrappedMsg := types.NewMsgWrappedBeginRedelegate(tc.req)

		_, err := msgSrvr.WrappedBeginRedelegate(wctx, wrappedMsg)
		require.NoError(t, err)

		resp, err := queryClient.EpochMsgs(wctx, &types.QueryEpochMsgsRequest{
			Pagination: &query.PageRequest{},
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Msgs))

		_, err = msgSrvr.WrappedBeginRedelegate(wctx, wrappedMsg)
		if tc.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
