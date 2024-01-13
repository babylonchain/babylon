package btccheckpoint

import (
	"context"
	"github.com/babylonchain/babylon/x/btccheckpoint/keeper"
)

// EndBlocker checks if during block execution btc light client head had been
// updated. If the head had been updated, status of all available checkpoints
// is checked to determine if any of them became confirmed/finalized/abandonded.
func EndBlocker(ctx context.Context, k keeper.Keeper) {
	if k.BtcLightClientUpdated(ctx) {
		k.OnTipChange(ctx)
	}
}
