package query

import (
	"context"

	monitortypes "github.com/babylonchain/babylon/x/monitor/types"
	"github.com/cosmos/cosmos-sdk/client"
)

// QueryMonitor queries the Monitor module of the Babylon node
// according to the given function
func (c *QueryClient) QueryMonitor(f func(ctx context.Context, queryClient monitortypes.QueryClient) error) error {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	clientCtx := client.Context{Client: c.RPCClient}
	queryClient := monitortypes.NewQueryClient(clientCtx)

	return f(ctx, queryClient)
}

// EndedEpochBTCHeight queries the tip height of BTC light client at epoch ends
func (c *QueryClient) EndedEpochBTCHeight(epochNum uint64) (*monitortypes.QueryEndedEpochBtcHeightResponse, error) {
	var resp *monitortypes.QueryEndedEpochBtcHeightResponse
	err := c.QueryMonitor(func(ctx context.Context, queryClient monitortypes.QueryClient) error {
		var err error
		req := &monitortypes.QueryEndedEpochBtcHeightRequest{
			EpochNum: epochNum,
		}
		resp, err = queryClient.EndedEpochBtcHeight(ctx, req)
		return err
	})

	return resp, err
}

// ReportedCheckpointBTCHeight queries the tip height of BTC light client when a given checkpoint is reported
func (c *QueryClient) ReportedCheckpointBTCHeight(hashStr string) (*monitortypes.QueryReportedCheckpointBtcHeightResponse, error) {
	var resp *monitortypes.QueryReportedCheckpointBtcHeightResponse
	err := c.QueryMonitor(func(ctx context.Context, queryClient monitortypes.QueryClient) error {
		var err error
		req := &monitortypes.QueryReportedCheckpointBtcHeightRequest{
			CkptHash: hashStr,
		}
		resp, err = queryClient.ReportedCheckpointBtcHeight(ctx, req)
		return err
	})

	return resp, err
}
