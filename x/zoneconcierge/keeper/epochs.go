package keeper

import (
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetLastSentSegment get last broadcasted btc light client segment
func (k Keeper) GetLastSentSegment(ctx sdk.Context) *types.BTCChainSegment {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.FinalizingBTCTipKey) {
		return nil
	}
	segmentBytes := store.Get(types.FinalizingBTCTipKey)
	var segment types.BTCChainSegment
	k.cdc.MustUnmarshal(segmentBytes, &segment)
	return &segment
}

// setLastSentSegment sets the last segment which was broadcasted to the other light clients
// called upon each AfterRawCheckpointFinalized hook invocation
func (k Keeper) setLastSentSegment(ctx sdk.Context, segment *types.BTCChainSegment) {
	store := ctx.KVStore(k.storeKey)
	segmentBytes := k.cdc.MustMarshal(segment)
	store.Set(types.FinalizingBTCTipKey, segmentBytes)
}

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
func (k Keeper) setFinalizedEpoch(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	store.Set(types.FinalizedEpochKey, epochNumberBytes)
}

func (k Keeper) GetEpoch(ctx sdk.Context) *epochingtypes.Epoch {
	return k.epochingKeeper.GetEpoch(ctx)
}
