package keeper_test

import (
	"bytes"
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func (suite *KeeperTestSuite) TestParamsQuery() {
	ctx, queryClient := suite.ctx, suite.queryClient
	req := types.QueryParamsRequest{}

	testCases := []struct {
		msg     string
		expPass bool
		params  types.Params
	}{
		{
			"default params",
			true,
			types.DefaultParams(),
		},
		{
			"wrong params",
			false,
			types.NewParams(777),
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			wctx := sdk.WrapSDKContext(ctx)
			resp, err := queryClient.Params(wctx, &req)
			suite.NoError(err)
			if tc.expPass {
				suite.Equal(&types.QueryParamsResponse{Params: tc.params}, resp)
			} else {
				suite.NotEqual(&types.QueryParamsResponse{Params: tc.params}, resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCurrentEpoch() {
	ctx, queryClient := suite.ctx, suite.queryClient
	req := types.QueryCurrentEpochRequest{}

	testCases := []struct {
		msg           string
		malleate      func()
		expPass       bool
		epochNumber   sdk.Uint
		epochBoundary sdk.Uint
	}{
		{
			"epoch 0",
			func() {},
			true,
			sdk.NewUint(0),
			sdk.NewUint(0),
		},
		{
			"epoch 1",
			func() {
				suite.keeper.IncEpochNumber(suite.ctx)
			},
			true,
			sdk.NewUint(1),
			sdk.NewUint(suite.keeper.GetParams(suite.ctx).EpochInterval * 1),
		},
		{
			"epoch 2",
			func() {
				suite.keeper.IncEpochNumber(suite.ctx)
			},
			true,
			sdk.NewUint(2),
			sdk.NewUint(suite.keeper.GetParams(suite.ctx).EpochInterval * 2),
		},
		{
			"reset to epoch 0",
			func() {
				suite.keeper.SetEpochNumber(suite.ctx, sdk.NewUint(0))
			},
			true,
			sdk.NewUint(0),
			sdk.NewUint(0),
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			wctx := sdk.WrapSDKContext(ctx)
			resp, err := queryClient.CurrentEpoch(wctx, &req)
			suite.NoError(err)
			if tc.expPass {
				suite.Equal(tc.epochNumber.Uint64(), resp.CurrentEpoch)
				suite.Equal(tc.epochBoundary.Uint64(), resp.EpochBoundary)
			} else {
				suite.NotEqual(tc.epochNumber.Uint64(), resp.CurrentEpoch)
				suite.NotEqual(tc.epochBoundary.Uint64(), resp.EpochBoundary)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEpochMsgs() {
	ctx, queryClient := suite.ctx, suite.queryClient

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		req       *types.QueryEpochMsgsRequest
		epochMsgs []*types.QueuedMessage
	}{
		{
			"empty epoch msgs",
			func() {},
			true,
			&types.QueryEpochMsgsRequest{
				Pagination: &query.PageRequest{
					Limit: 100,
				},
			},
			[]*types.QueuedMessage{},
		},
		{
			"non-exist epoch msg",
			func() {},
			false,
			&types.QueryEpochMsgsRequest{
				Pagination: &query.PageRequest{
					Limit: 100,
				},
			},
			[]*types.QueuedMessage{
				{TxId: []byte{0x01}},
			},
		},
		{
			"newly inserted epoch msg",
			func() {
				msg := types.QueuedMessage{
					TxId: []byte{0x01},
				}
				suite.keeper.EnqueueMsg(suite.ctx, msg)
			},
			true,
			&types.QueryEpochMsgsRequest{
				Pagination: &query.PageRequest{
					Limit: 100,
				},
			},
			[]*types.QueuedMessage{
				{TxId: []byte{0x01}},
			},
		},
		{
			"cleared epoch msg",
			func() {
				suite.keeper.ClearEpochMsgs(suite.ctx)
			},
			true,
			&types.QueryEpochMsgsRequest{
				Pagination: &query.PageRequest{
					Limit: 100,
				},
			},
			[]*types.QueuedMessage{},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			wctx := sdk.WrapSDKContext(ctx)
			resp, err := queryClient.EpochMsgs(wctx, tc.req)
			suite.NoError(err)
			if tc.expPass {
				suite.Equal(len(tc.epochMsgs), len(resp.Msgs))
				suite.Equal(uint64(len(tc.epochMsgs)), suite.keeper.GetQueueLength(suite.ctx).Uint64())
				for idx := range tc.epochMsgs {
					suite.Equal(tc.epochMsgs[idx].MsgId, resp.Msgs[idx].MsgId)
					suite.Equal(tc.epochMsgs[idx].TxId, resp.Msgs[idx].TxId)
				}
			} else {
				if len(tc.epochMsgs) != len(resp.Msgs) {
					suite.T().Skip()
				}
				eq := true
				for idx := range tc.epochMsgs {
					if !bytes.Equal(tc.epochMsgs[idx].MsgId, resp.Msgs[idx].MsgId) || !bytes.Equal(tc.epochMsgs[idx].TxId, resp.Msgs[idx].TxId) {
						eq = false
					}
				}
				suite.False(eq)
			}
		})
	}
}
