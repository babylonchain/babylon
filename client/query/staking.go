package query

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// QueryStaking queries the Staking module of the Babylon node
// according to the given function
func (c *QueryClient) QueryStaking(f func(ctx context.Context, queryClient stakingtypes.QueryClient) error) error {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	clientCtx := client.Context{Client: c.RPCClient}
	queryClient := stakingtypes.NewQueryClient(clientCtx)

	return f(ctx, queryClient)
}

// StakingParams queries btccheckpoint module's parameters via ChainClient
func (c *QueryClient) StakingParams() (*stakingtypes.QueryParamsResponse, error) {
	var resp *stakingtypes.QueryParamsResponse
	err := c.QueryStaking(func(ctx context.Context, queryClient stakingtypes.QueryClient) error {
		var err error
		req := &stakingtypes.QueryParamsRequest{}
		resp, err = queryClient.Params(ctx, req)
		return err
	})

	return resp, err
}
