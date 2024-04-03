package query

import (
	"context"

	incentivetypes "github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/client"
)

// QueryIncentive queries the Incentive module of the Babylon node
func (c *QueryClient) QueryIncentive(f func(ctx context.Context, queryClient incentivetypes.QueryClient) error) error {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	clientCtx := client.Context{Client: c.RPCClient}
	queryClient := incentivetypes.NewQueryClient(clientCtx)

	return f(ctx, queryClient)
}

// RewardGauges queries the Incentive module to get all reward gauges
func (c *QueryClient) RewardGauges(address string) (*incentivetypes.QueryRewardGaugesResponse, error) {
	var resp *incentivetypes.QueryRewardGaugesResponse
	err := c.QueryIncentive(func(ctx context.Context, queryClient incentivetypes.QueryClient) error {
		var err error
		req := &incentivetypes.QueryRewardGaugesRequest{
			Address: address,
		}
		resp, err = queryClient.RewardGauges(ctx, req)
		return err
	})

	return resp, err
}
