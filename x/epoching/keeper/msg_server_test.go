package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	testhelper "github.com/babylonchain/babylon/testutil/helper"
	"github.com/babylonchain/babylon/x/epoching/types"
)

// TODO (fuzz tests): replace the following tests with fuzz ones
func TestMsgWrappedDelegate(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	helper := testhelper.NewHelper(t)
	msgSrvr := helper.MsgSrvr
	// enter 1st epoch, in which BBN starts handling validator-related msgs
	ctx, err := helper.ApplyEmptyBlockWithVoteExtension(r)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		req       *stakingtypes.MsgDelegate
		expectErr bool
	}{
		{
			"empty wrapped msg",
			&stakingtypes.MsgDelegate{},
			true,
		},
	}
	for _, tc := range testCases {
		wrappedMsg := types.NewMsgWrappedDelegate(tc.req)
		_, err := msgSrvr.WrappedDelegate(ctx, wrappedMsg)
		if tc.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestMsgWrappedUndelegate(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	helper := testhelper.NewHelper(t)
	msgSrvr := helper.MsgSrvr
	// enter 1st epoch, in which BBN starts handling validator-related msgs
	ctx, err := helper.ApplyEmptyBlockWithVoteExtension(r)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		req       *stakingtypes.MsgUndelegate
		expectErr bool
	}{
		{
			"empty wrapped msg",
			&stakingtypes.MsgUndelegate{},
			true,
		},
	}
	for _, tc := range testCases {
		wrappedMsg := types.NewMsgWrappedUndelegate(tc.req)
		_, err := msgSrvr.WrappedUndelegate(ctx, wrappedMsg)
		if tc.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestMsgWrappedBeginRedelegate(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	helper := testhelper.NewHelper(t)
	msgSrvr := helper.MsgSrvr
	// enter 1st epoch, in which BBN starts handling validator-related msgs
	ctx, err := helper.ApplyEmptyBlockWithVoteExtension(r)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		req       *stakingtypes.MsgBeginRedelegate
		expectErr bool
	}{
		{
			"empty wrapped msg",
			&stakingtypes.MsgBeginRedelegate{},
			true,
		},
	}
	for _, tc := range testCases {
		wrappedMsg := types.NewMsgWrappedBeginRedelegate(tc.req)

		_, err := msgSrvr.WrappedBeginRedelegate(ctx, wrappedMsg)
		if tc.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestMsgWrappedCancelUnbondingDelegation(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	helper := testhelper.NewHelper(t)
	msgSrvr := helper.MsgSrvr
	// enter 1st epoch, in which BBN starts handling validator-related msgs
	ctx, err := helper.ApplyEmptyBlockWithVoteExtension(r)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		req       *stakingtypes.MsgCancelUnbondingDelegation
		expectErr bool
	}{
		{
			"empty wrapped msg",
			&stakingtypes.MsgCancelUnbondingDelegation{},
			true,
		},
	}
	for _, tc := range testCases {
		wrappedMsg := types.NewMsgWrappedCancelUnbondingDelegation(tc.req)

		_, err := msgSrvr.WrappedCancelUnbondingDelegation(ctx, wrappedMsg)
		if tc.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
