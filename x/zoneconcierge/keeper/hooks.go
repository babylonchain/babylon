package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
)

type Hooks struct {
	k Keeper
}

// ensures Hooks implements ClientHooks interfaces
var _ checkpointingtypes.CheckpointingHooks = Hooks{}
var _ epochingtypes.EpochingHooks = Hooks{}

func (k Keeper) Hooks() Hooks { return Hooks{k} }

func (h Hooks) AfterEpochEnds(ctx context.Context, epoch uint64) {
	// upon an epoch has ended, index the current chain info for each CZ
	// TODO: do this together when epoch is sealed?
	for _, chainID := range h.k.GetAllChainIDs(ctx) {
		h.k.recordEpochChainInfo(ctx, chainID, epoch)
	}
}

func (h Hooks) AfterRawCheckpointSealed(ctx context.Context, epoch uint64) error {
	// upon a raw checkpoint is sealed, index the current chain info for each consumer,
	// and generate/save the proof that the epoch is sealed
	h.k.recordEpochChainInfoProofs(ctx, epoch)
	h.k.recordSealedEpochProof(ctx, epoch)
	return nil
}

// AfterRawCheckpointFinalized is triggered upon an epoch has been finalised
func (h Hooks) AfterRawCheckpointFinalized(ctx context.Context, epoch uint64) error {
	headersToBroadcast := h.k.getHeadersToBroadcast(ctx)

	// send BTC timestamp to all open channels with ZoneConcierge
	// TODO: BroadcastBTCTimestamps is non-deterministic due to generating proofs
	// which are affected by pruning. Re-enable after improving BroadcastBTCTimestamps
	// methods
	// h.k.BroadcastBTCTimestamps(ctx, epoch, headersToBroadcast)

	// Update the last broadcasted segment
	h.k.setLastSentSegment(ctx, &types.BTCChainSegment{
		BtcHeaders: headersToBroadcast,
	})
	return nil
}

// Other unused hooks

func (h Hooks) AfterBlsKeyRegistered(ctx context.Context, valAddr sdk.ValAddress) error { return nil }
func (h Hooks) AfterRawCheckpointConfirmed(ctx context.Context, epoch uint64) error     { return nil }
func (h Hooks) AfterRawCheckpointForgotten(ctx context.Context, ckpt *checkpointingtypes.RawCheckpoint) error {
	return nil
}
func (h Hooks) AfterRawCheckpointBlsSigVerified(ctx context.Context, ckpt *checkpointingtypes.RawCheckpoint) error {
	return nil
}

func (h Hooks) AfterEpochBegins(ctx context.Context, epoch uint64)                          {}
func (h Hooks) BeforeSlashThreshold(ctx context.Context, valSet epochingtypes.ValidatorSet) {}
