package keeper_test

import (
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// TODO: fuzz tests
// 1. Generate some random params and perhaps a flag
// 2. If the flag is set, set the params it in the keeper
// 3. Send the query to get the current state
// 4. If the flag was set, verify that the generated parameter was returned; otherwise verify that the default params are returned
func (suite *KeeperTestSuite) TestParamsQuery() {
	ctx, queryClient := suite.ctx, suite.queryClient
	req := types.QueryParamsRequest{}

	testCases := []struct {
		msg    string
		params types.Params
	}{
		{
			"default params",
			types.DefaultParams(),
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			wctx := sdk.WrapSDKContext(ctx)
			resp, err := queryClient.Params(wctx, &req)
			suite.NoError(err)
			suite.Equal(&types.QueryParamsResponse{Params: tc.params}, resp)
		})
	}
}

// TODO: fuzz tests
//  generate these steps of increments/sets randomly and calculate the expected state in the test on an idealised model, and check the real implementation against that.
func (suite *KeeperTestSuite) TestCurrentEpoch() {
	ctx, queryClient := suite.ctx, suite.queryClient
	req := types.QueryCurrentEpochRequest{}

	testCases := []struct {
		msg           string
		malleate      func()
		epochNumber   sdk.Uint
		epochBoundary sdk.Uint
	}{
		{
			"epoch 0",
			func() {},
			sdk.NewUint(0),
			sdk.NewUint(0),
		},
		{
			"epoch 1",
			func() {
				suite.keeper.IncEpochNumber(suite.ctx)
			},
			sdk.NewUint(1),
			sdk.NewUint(suite.keeper.GetParams(suite.ctx).EpochInterval * 1),
		},
		{
			"epoch 2",
			func() {
				suite.keeper.IncEpochNumber(suite.ctx)
			},
			sdk.NewUint(2),
			sdk.NewUint(suite.keeper.GetParams(suite.ctx).EpochInterval * 2),
		},
		{
			"reset to epoch 0",
			func() {
				suite.keeper.SetEpochNumber(suite.ctx, sdk.NewUint(0))
			},
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
			suite.Equal(tc.epochNumber.Uint64(), resp.CurrentEpoch)
			suite.Equal(tc.epochBoundary.Uint64(), resp.EpochBoundary)
		})
	}
}

// TODO: fuzz tests
// randomly generate the limit in the request, to check that it is respected. Something like this:
func (suite *KeeperTestSuite) TestEpochMsgs() {
	ctx, queryClient := suite.ctx, suite.queryClient

	testCases := []struct {
		msg       string
		malleate  func()
		req       *types.QueryEpochMsgsRequest
		epochMsgs []*types.QueuedMessage
	}{
		{
			"empty epoch msgs",
			func() {},
			&types.QueryEpochMsgsRequest{
				Pagination: &query.PageRequest{
					Limit: 100,
				},
			},
			[]*types.QueuedMessage{},
		},
		{
			"newly inserted epoch msg",
			func() {
				msg := types.QueuedMessage{
					TxId: []byte{0x01},
				}
				suite.keeper.EnqueueMsg(suite.ctx, msg)
			},
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
			suite.Equal(len(tc.epochMsgs), len(resp.Msgs))
			suite.Equal(uint64(len(tc.epochMsgs)), suite.keeper.GetQueueLength(suite.ctx).Uint64())
			for idx := range tc.epochMsgs {
				suite.Equal(tc.epochMsgs[idx].MsgId, resp.Msgs[idx].MsgId)
				suite.Equal(tc.epochMsgs[idx].TxId, resp.Msgs[idx].TxId)
			}
		})
	}
}
