package types

import (
	"context"

	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
)

// EpochingKeeper defines the expected interface needed to retrieve epoch info
type EpochingKeeper interface {
	GetEpoch(ctx context.Context) *epochingtypes.Epoch
	EnqueueMsg(ctx context.Context, msg epochingtypes.QueuedMessage)
	GetValidatorSet(ctx context.Context, epochNumer uint64) epochingtypes.ValidatorSet
	GetTotalVotingPower(ctx context.Context, epochNumber uint64) int64
	CheckMsgCreateValidator(ctx context.Context, msg *stakingtypes.MsgCreateValidator) error
	GetPubKeyByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (cmtprotocrypto.PublicKey, error)
}

// Event Hooks
// These can be utilized to communicate between a checkpointing keeper and another
// keeper which must take particular actions when raw checkpoints change
// state. The second keeper must implement this interface, which then the
// checkpointing keeper can call.

// CheckpointingHooks event hooks for raw checkpoint object (noalias)
type CheckpointingHooks interface {
	AfterBlsKeyRegistered(ctx context.Context, valAddr sdk.ValAddress) error         // Must be called when a BLS key is registered
	AfterRawCheckpointSealed(ctx context.Context, epoch uint64) error                // Must be called when a raw checkpoint is SEALED
	AfterRawCheckpointConfirmed(ctx context.Context, epoch uint64) error             // Must be called when a raw checkpoint is CONFIRMED
	AfterRawCheckpointForgotten(ctx context.Context, ckpt *RawCheckpoint) error      // Must be called when a raw checkpoint is FORGOTTEN
	AfterRawCheckpointFinalized(ctx context.Context, epoch uint64) error             // Must be called when a raw checkpoint is FINALIZED
	AfterRawCheckpointBlsSigVerified(ctx context.Context, ckpt *RawCheckpoint) error // Must be called when a raw checkpoint's multi-sig is verified
}
