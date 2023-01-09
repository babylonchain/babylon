package keeper

import (
	"fmt"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetInfoForFinalizedEpoch gets the raw checkpoint and the associated best submission of a finalised epoch
// CONTRACT: the function can only take an epoch that has already been finalised as input
func (k Keeper) GetInfoForFinalizedEpoch(ctx sdk.Context, epochNumber uint64) (*checkpointingtypes.RawCheckpoint, *btcctypes.SubmissionKey, error) {
	// find the btc checkpoint tx index of this epoch
	ed := k.btccKeeper.GetEpochData(ctx, epochNumber)
	if ed.Status != btcctypes.Finalized {
		err := fmt.Errorf("epoch %d should have been finalized, but is in status %s", epochNumber, ed.Status.String())
		panic(err) // this can only be a programming error
	}
	if len(ed.Key) == 0 {
		err := fmt.Errorf("finalized epoch %d should have at least 1 checkpoint submission", epochNumber)
		panic(err) // this can only be a programming error
	}
	bestSubmissionKey := ed.Key[0] // index of checkpoint tx on BTC

	// get raw checkpoint of this epoch
	rawCheckpointBytes := ed.RawCheckpoint
	rawCheckpoint, err := checkpointingtypes.FromBTCCkptBytesToRawCkpt(rawCheckpointBytes)
	if err != nil {
		return nil, bestSubmissionKey, err
	}
	return rawCheckpoint, bestSubmissionKey, nil
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
