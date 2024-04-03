package query

import (
	"context"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
)

// QueryBTCCheckpoint queries the BTCCheckpoint module of the Babylon node
// according to the given function
func (c *QueryClient) QueryBTCCheckpoint(f func(ctx context.Context, queryClient btcctypes.QueryClient) error) error {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	clientCtx := client.Context{Client: c.RPCClient}
	queryClient := btcctypes.NewQueryClient(clientCtx)

	return f(ctx, queryClient)
}

// BTCCheckpointParams queries btccheckpoint module's parameters via ChainClient
func (c *QueryClient) BTCCheckpointParams() (*btcctypes.QueryParamsResponse, error) {
	var resp *btcctypes.QueryParamsResponse
	err := c.QueryBTCCheckpoint(func(ctx context.Context, queryClient btcctypes.QueryClient) error {
		var err error
		req := &btcctypes.QueryParamsRequest{}
		resp, err = queryClient.Params(ctx, req)
		return err
	})

	return resp, err
}

// BTCCheckpointInfo queries btccheckpoint module for the Bitcoin position of an epoch
func (c *QueryClient) BTCCheckpointInfo(epochNumber uint64) (*btcctypes.QueryBtcCheckpointInfoResponse, error) {
	var resp *btcctypes.QueryBtcCheckpointInfoResponse
	err := c.QueryBTCCheckpoint(func(ctx context.Context, queryClient btcctypes.QueryClient) error {
		var err error
		req := &btcctypes.QueryBtcCheckpointInfoRequest{
			EpochNum: epochNumber,
		}
		resp, err = queryClient.BtcCheckpointInfo(ctx, req)
		return err
	})

	return resp, err
}

// BTCCheckpointsInfo queries btccheckpoint module for the Bitcoin position of an epoch range
func (c *QueryClient) BTCCheckpointsInfo(pagination *sdkquerytypes.PageRequest) (*btcctypes.QueryBtcCheckpointsInfoResponse, error) {
	var resp *btcctypes.QueryBtcCheckpointsInfoResponse
	err := c.QueryBTCCheckpoint(func(ctx context.Context, queryClient btcctypes.QueryClient) error {
		var err error
		req := &btcctypes.QueryBtcCheckpointsInfoRequest{
			Pagination: pagination,
		}
		resp, err = queryClient.BtcCheckpointsInfo(ctx, req)
		return err
	})
	return resp, err
}
