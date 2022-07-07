package keeper_test

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO: check if the msg is indeed queued
func (suite *KeeperTestSuite) TestMsgWrappedDelegate() {
	testCases := []struct {
		name      string
		req       func() *types.MsgWrappedDelegate
		expectErr bool
	}{
		{
			"MsgWrappedDelegate",
			func() *types.MsgWrappedDelegate {
				return &types.MsgWrappedDelegate{
					Msg: &stakingtypes.MsgDelegate{},
				}
			},
			false,
		},
	}
	for _, tc := range testCases {
		wctx := sdk.WrapSDKContext(suite.ctx)
		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.WrappedDelegate(wctx, tc.req())
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgWrappedUndelegate() {
	testCases := []struct {
		name      string
		req       func() *types.MsgWrappedUndelegate
		expectErr bool
	}{
		{
			"MsgWrappedDelegate",
			func() *types.MsgWrappedUndelegate {
				return &types.MsgWrappedUndelegate{
					Msg: &stakingtypes.MsgUndelegate{},
				}
			},
			false,
		},
	}
	for _, tc := range testCases {
		wctx := sdk.WrapSDKContext(suite.ctx)
		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.WrappedUndelegate(wctx, tc.req())
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgWrappedBeginRedelegate() {
	testCases := []struct {
		name      string
		req       func() *types.MsgWrappedBeginRedelegate
		expectErr bool
	}{
		{
			"MsgWrappedDelegate",
			func() *types.MsgWrappedBeginRedelegate {
				return &types.MsgWrappedBeginRedelegate{
					Msg: &stakingtypes.MsgBeginRedelegate{},
				}
			},
			false,
		},
	}
	for _, tc := range testCases {
		wctx := sdk.WrapSDKContext(suite.ctx)
		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.WrappedBeginRedelegate(wctx, tc.req())
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
