package keeper

import (
	"context"

	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
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

func (k Keeper) GetLastFinalizedEpoch(ctx context.Context) (uint64, error) {
	return k.checkpointingKeeper.GetLastFinalizedEpoch(ctx)
}

func (k Keeper) GetEpoch(ctx context.Context) *epochingtypes.Epoch {
	return k.epochingKeeper.GetEpoch(ctx)
}
