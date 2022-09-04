package keeper_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// TODO (fuzz tests): replace the following tests with fuzz ones
func TestMsgWrappedDelegate(t *testing.T) {
	helper := testepoching.NewHelper(t)
	msgSrvr, queryClient := helper.MsgSrvr, helper.QueryClient
	// enter 1st epoch, in which BBN starts handling validator-related msgs
	ctx := helper.GenAndApplyEmptyBlock()
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
			EpochNum:   uint64(1),
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
	helper := testepoching.NewHelper(t)
	msgSrvr, queryClient := helper.MsgSrvr, helper.QueryClient
	// enter 1st epoch, in which BBN starts handling validator-related msgs
	ctx := helper.GenAndApplyEmptyBlock()
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
			EpochNum:   uint64(1),
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
	helper := testepoching.NewHelper(t)
	msgSrvr, queryClient := helper.MsgSrvr, helper.QueryClient
	// enter 1st epoch, in which BBN starts handling validator-related msgs
	ctx := helper.GenAndApplyEmptyBlock()
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
			EpochNum:   uint64(1),
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
