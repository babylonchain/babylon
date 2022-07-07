package keeper_test

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO: check if the msg is indeed queued
func (suite *KeeperTestSuite) TestMsgWrappedDelegate() {
	testCases := []struct {
		name      string
		req       *types.MsgWrappedDelegate
		expectErr bool
	}{
		{
			"MsgWrappedDelegate",
			&types.MsgWrappedDelegate{
				Msg: &stakingtypes.MsgDelegate{},
			},
			false,
		},
	}
	for _, tc := range testCases {
		wctx := sdk.WrapSDKContext(suite.ctx)
		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.WrappedDelegate(wctx, tc.req)
			suite.Require().NoError(err)

			resp, err := suite.queryClient.EpochMsgs(wctx, &types.QueryEpochMsgsRequest{
				Pagination: &query.PageRequest{},
			})
			suite.Require().NoError(err)
			suite.Require().Equal(1, len(resp.Msgs))

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
		req       *types.MsgWrappedUndelegate
		expectErr bool
	}{
		{
			"MsgWrappedDelegate",
			&types.MsgWrappedUndelegate{
				Msg: &stakingtypes.MsgUndelegate{},
			},
			false,
		},
	}
	for _, tc := range testCases {
		wctx := sdk.WrapSDKContext(suite.ctx)
		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.WrappedUndelegate(wctx, tc.req)
			suite.Require().NoError(err)

			resp, err := suite.queryClient.EpochMsgs(wctx, &types.QueryEpochMsgsRequest{
				Pagination: &query.PageRequest{},
			})
			suite.Require().NoError(err)
			suite.Require().Equal(1, len(resp.Msgs))

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
		req       *types.MsgWrappedBeginRedelegate
		expectErr bool
	}{
		{
			"MsgWrappedDelegate",
			&types.MsgWrappedBeginRedelegate{
				Msg: &stakingtypes.MsgBeginRedelegate{},
			},
			false,
		},
	}
	for _, tc := range testCases {
		wctx := sdk.WrapSDKContext(suite.ctx)
		_, err := suite.msgSrvr.WrappedBeginRedelegate(wctx, tc.req)
		suite.Require().NoError(err)

		resp, err := suite.queryClient.EpochMsgs(wctx, &types.QueryEpochMsgsRequest{
			Pagination: &query.PageRequest{},
		})
		suite.Require().NoError(err)
		suite.Require().Equal(1, len(resp.Msgs))

		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.WrappedBeginRedelegate(wctx, tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
