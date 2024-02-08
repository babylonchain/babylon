package keeper

import (
	"context"

	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetLastSentSegment get last broadcasted btc light client segment
func (k Keeper) GetLastSentSegment(ctx context.Context) *types.BTCChainSegment {
	store := k.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.LastSentBTCSegmentKey)
	if err != nil {
		panic(err)
	}
	if !has {
		return nil
	}
	segmentBytes, err := store.Get(types.LastSentBTCSegmentKey)
	if err != nil {
		panic(err)
	}
	var segment types.BTCChainSegment
	k.cdc.MustUnmarshal(segmentBytes, &segment)
	return &segment
}

// setLastSentSegment sets the last segment which was broadcasted to the other light clients
// called upon each AfterRawCheckpointFinalized hook invocation
func (k Keeper) setLastSentSegment(ctx context.Context, segment *types.BTCChainSegment) {
	store := k.storeService.OpenKVStore(ctx)
	segmentBytes := k.cdc.MustMarshal(segment)
	if err := store.Set(types.LastSentBTCSegmentKey, segmentBytes); err != nil {
		panic(err)
	}
}

// GetFinalizedEpoch gets the last finalised epoch
// used upon querying the last BTC-finalised chain info for CZs
func (k Keeper) GetFinalizedEpoch(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.FinalizedEpochKey)
	if err != nil {
		panic(err)
	}
	if !has {
		return 0, types.ErrFinalizedEpochNotFound
	}
	epochNumberBytes, err := store.Get(types.FinalizedEpochKey)
	if err != nil {
		panic(err)
	}
	return sdk.BigEndianToUint64(epochNumberBytes), nil
}

// setFinalizedEpoch sets the last finalised epoch
// called upon each AfterRawCheckpointFinalized hook invocation
func (k Keeper) setFinalizedEpoch(ctx context.Context, epochNumber uint64) {
	store := k.storeService.OpenKVStore(ctx)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	if err := store.Set(types.FinalizedEpochKey, epochNumberBytes); err != nil {
		panic(err)
	}
}

func (k Keeper) GetEpoch(ctx context.Context) *epochingtypes.Epoch {
	return k.epochingKeeper.GetEpoch(ctx)
}
