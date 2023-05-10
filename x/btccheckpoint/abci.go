package btccheckpoint

import (
	"github.com/babylonchain/babylon/x/btccheckpoint/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker checks if during block execution btc light client head had been
// updated. If the head had been updated, status of all available checkpoints
// is checked to determine if any of them became confirmed/finalized/abandonded.
func EndBlocker(ctx sdk.Context, k keeper.Keeper, req abci.RequestEndBlock) {
	if k.BtcLightClientUpdated(ctx) {
		k.OnTipChange(ctx)
	}
}
