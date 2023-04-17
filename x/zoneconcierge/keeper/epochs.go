package keeper

import (
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetFinalizedEpoch gets the last finalised epoch
// used upon querying the last BTC-finalised chain info for CZs
func (k Keeper) GetFinalizedEpoch(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.FinalizedEpochKey) {
		return 0, types.ErrFinalizedEpochNotFound
	}
	epochNumberBytes := store.Get(types.FinalizedEpochKey)
	return sdk.BigEndianToUint64(epochNumberBytes), nil
}

// setFinalizedEpoch sets the last finalised epoch
// called upon each AfterRawCheckpointFinalized hook invocation
// NOTE: Public for test purposes
func (k Keeper) SetFinalizedEpoch(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	store.Set(types.FinalizedEpochKey, epochNumberBytes)
}

func (k Keeper) GetEpoch(ctx sdk.Context) *epochingtypes.Epoch {
	return k.epochingKeeper.GetEpoch(ctx)
}
