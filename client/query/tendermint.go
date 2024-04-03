package query

import (
	"context"
	"strings"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
)

// GetStatus returns the status of the tendermint node
func (c *QueryClient) GetStatus() (*coretypes.ResultStatus, error) {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	return c.RPCClient.Status(ctx)
}

// GetBlock returns the tendermint block at a specific height
func (c *QueryClient) GetBlock(height int64) (*coretypes.ResultBlock, error) {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	return c.RPCClient.Block(ctx, &height)
}

// BlockSearch searches for blocks satisfying the events specified on the events list
func (c *QueryClient) BlockSearch(events []string, page *int, perPage *int, orderBy string) (*coretypes.ResultBlockSearch, error) {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	return c.RPCClient.BlockSearch(ctx, strings.Join(events, " AND "), page, perPage, orderBy)
}

// TxSearch searches for transactions satisfying the events specified on the events list
func (c *QueryClient) TxSearch(events []string, prove bool, page *int, perPage *int, orderBy string) (*coretypes.ResultTxSearch, error) {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	return c.RPCClient.TxSearch(ctx, strings.Join(events, " AND "), prove, page, perPage, orderBy)
}

// GetTx returns the transaction with the specified hash
func (c *QueryClient) GetTx(hash []byte) (*coretypes.ResultTx, error) {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	return c.RPCClient.Tx(ctx, hash, false)
}

func (c *QueryClient) Subscribe(subscriber, query string, outCapacity ...int) (out <-chan coretypes.ResultEvent, err error) {
	return c.RPCClient.Subscribe(context.Background(), subscriber, query, outCapacity...)
}

func (c *QueryClient) Unsubscribe(subscriber, query string) error {
	return c.RPCClient.Unsubscribe(context.Background(), subscriber, query)
}

func (c *QueryClient) UnsubscribeAll(subscriber string) error {
	return c.RPCClient.UnsubscribeAll(context.Background(), subscriber)
}
