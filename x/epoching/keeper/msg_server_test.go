package keeper_test

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO (fuzz tests): replace the following tests with fuzz ones
func (suite *KeeperTestSuite) TestMsgWrappedDelegate() {
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
		wctx := sdk.WrapSDKContext(suite.ctx)
		suite.Run(tc.name, func() {
			wrappedMsg := types.NewMsgWrappedDelegate(tc.req)
			_, err := suite.msgSrvr.WrappedDelegate(wctx, wrappedMsg)
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
		wctx := sdk.WrapSDKContext(suite.ctx)
		suite.Run(tc.name, func() {
			wrappedMsg := types.NewMsgWrappedUndelegate(tc.req)
			_, err := suite.msgSrvr.WrappedUndelegate(wctx, wrappedMsg)
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
		wctx := sdk.WrapSDKContext(suite.ctx)
		wrappedMsg := types.NewMsgWrappedBeginRedelegate(tc.req)

		_, err := suite.msgSrvr.WrappedBeginRedelegate(wctx, wrappedMsg)
		suite.Require().NoError(err)

		resp, err := suite.queryClient.EpochMsgs(wctx, &types.QueryEpochMsgsRequest{
			Pagination: &query.PageRequest{},
		})
		suite.Require().NoError(err)
		suite.Require().Equal(1, len(resp.Msgs))

		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.WrappedBeginRedelegate(wctx, wrappedMsg)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
