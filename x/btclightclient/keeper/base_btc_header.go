package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/btclightclient/types"
)

func (k Keeper) GetBaseBTCHeader(ctx context.Context) *types.BTCHeaderInfo {
	return k.headersState(ctx).BaseHeader()
}

// SetBaseBTCHeader checks whether a base BTC header exist, if not inserts it into storage
func (k Keeper) SetBaseBTCHeader(ctx context.Context, baseBTCHeader types.BTCHeaderInfo) {
	existingHeader := k.headersState(ctx).BaseHeader()
	if existingHeader != nil {
		panic("A base BTC Header has already been set")
	}
	k.headersState(ctx).insertHeader(&baseBTCHeader)
}
