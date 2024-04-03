package query

import (
	"context"

	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
)

// QueryEpoching queries the Epoching module of the Babylon node
// according to the given function
func (c *QueryClient) QueryEpoching(f func(ctx context.Context, queryClient epochingtypes.QueryClient) error) error {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	clientCtx := client.Context{Client: c.RPCClient}
	queryClient := epochingtypes.NewQueryClient(clientCtx)

	return f(ctx, queryClient)
}

// EpochingParams queries epoching module's parameters via ChainClient
func (c *QueryClient) EpochingParams() (*epochingtypes.QueryParamsResponse, error) {
	var resp *epochingtypes.QueryParamsResponse
	err := c.QueryEpoching(func(ctx context.Context, queryClient epochingtypes.QueryClient) error {
		var err error
		req := &epochingtypes.QueryParamsRequest{}
		resp, err = queryClient.Params(ctx, req)
		return err
	})

	return resp, err
}

// CurrentEpoch queries the current epoch number via ChainClient
func (c *QueryClient) CurrentEpoch() (*epochingtypes.QueryCurrentEpochResponse, error) {
	var resp *epochingtypes.QueryCurrentEpochResponse
	err := c.QueryEpoching(func(ctx context.Context, queryClient epochingtypes.QueryClient) error {
		var err error
		req := &epochingtypes.QueryCurrentEpochRequest{}
		resp, err = queryClient.CurrentEpoch(ctx, req)
		return err
	})

	return resp, err
}

// EpochsInfo queries the epoching module for the maintained epochs
func (c *QueryClient) EpochsInfo(pagination *sdkquerytypes.PageRequest) (*epochingtypes.QueryEpochsInfoResponse, error) {
	var resp *epochingtypes.QueryEpochsInfoResponse
	err := c.QueryEpoching(func(ctx context.Context, queryClient epochingtypes.QueryClient) error {
		var err error
		req := &epochingtypes.QueryEpochsInfoRequest{
			Pagination: pagination,
		}
		resp, err = queryClient.EpochsInfo(ctx, req)
		return err
	})

	return resp, err
}

// LatestEpochMsgs queries the epoching module for the latest messages maintained in its delayed
// staking queue until a specified endEpoch.
func (c *QueryClient) LatestEpochMsgs(endEpoch uint64, epochCount uint64, pagination *sdkquerytypes.PageRequest) (*epochingtypes.QueryLatestEpochMsgsResponse, error) {
	var resp *epochingtypes.QueryLatestEpochMsgsResponse
	err := c.QueryEpoching(func(ctx context.Context, queryClient epochingtypes.QueryClient) error {
		var err error
		req := &epochingtypes.QueryLatestEpochMsgsRequest{
			EndEpoch:   endEpoch,
			EpochCount: epochCount,
			Pagination: pagination,
		}
		resp, err = queryClient.LatestEpochMsgs(ctx, req)
		return err
	})

	return resp, err
}

// DelegationLifecycle queries the epoching module for the lifecycle of a delegator.
func (c *QueryClient) DelegationLifecycle(delegator string) (*epochingtypes.QueryDelegationLifecycleResponse, error) {
	var resp *epochingtypes.QueryDelegationLifecycleResponse
	err := c.QueryEpoching(func(ctx context.Context, queryClient epochingtypes.QueryClient) error {
		var err error
		req := &epochingtypes.QueryDelegationLifecycleRequest{
			DelAddr: delegator,
		}
		resp, err = queryClient.DelegationLifecycle(ctx, req)
		return err
	})

	return resp, err
}
