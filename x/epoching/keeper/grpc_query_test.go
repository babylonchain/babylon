package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/epoching/types"
)

func (suite *KeeperTestSuite) TestParamsQuery() {
	ctx, queryClient := suite.ctx, suite.queryClient
	var req *types.QueryParamsRequest

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		params   types.Params
	}{
		{
			"default params",
			func() {
				req = &types.QueryParamsRequest{}
			},
			true,
			types.DefaultParams(),
		},
		{
			"wrong params",
			func() {
				req = &types.QueryParamsRequest{}
			},
			false,
			types.NewParams(777),
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			wctx := sdk.WrapSDKContext(ctx)
			resp, err := queryClient.Params(wctx, req)
			if tc.expPass {
				suite.NoError(err)
				suite.Equal(&types.QueryParamsResponse{Params: tc.params}, resp)
			} else {
				suite.NotEqual(&types.QueryParamsResponse{Params: tc.params}, resp)
			}
		})
	}
}
