package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/runtime"
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

func (k Keeper) GetLastFinalizedEpoch(ctx context.Context) uint64 {
	return k.checkpointingKeeper.GetLastFinalizedEpoch(ctx)
}

func (k Keeper) GetEpoch(ctx context.Context) *epochingtypes.Epoch {
	return k.epochingKeeper.GetEpoch(ctx)
}

func (k Keeper) recordSealedEpochProof(ctx context.Context, epochNum uint64) {
	// proof that the epoch is sealed
	proofEpochSealed, err := k.ProveEpochSealed(ctx, epochNum)
	if err != nil {
		panic(err) // only programming error
	}

	store := k.sealedEpochProofStore(ctx)
	store.Set(sdk.Uint64ToBigEndian(epochNum), k.cdc.MustMarshal(proofEpochSealed))
}

func (k Keeper) getSealedEpochProof(ctx context.Context, epochNum uint64) *types.ProofEpochSealed {
	store := k.sealedEpochProofStore(ctx)
	proofBytes := store.Get(sdk.Uint64ToBigEndian(epochNum))
	if len(proofBytes) == 0 {
		return nil
	}
	var proof types.ProofEpochSealed
	k.cdc.MustUnmarshal(proofBytes, &proof)
	return &proof
}

// sealedEpochProofStore stores the proof that each epoch is sealed
// prefix: SealedEpochProofKey
// key: epochNumber
// value: ChainInfoWithProof
func (k Keeper) sealedEpochProofStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.SealedEpochProofKey)
}
