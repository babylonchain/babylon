package types

import (
	"context"
	txformat "github.com/babylonchain/babylon/btctxformatter"
	bbn "github.com/babylonchain/babylon/types"
)

type BTCLightClientKeeper interface {
	// BlockHeight should validate if header with given hash is valid and if it is
	// part of known chain. In case this is true it should return this block height
	// in case this is false it should return error
	BlockHeight(ctx context.Context, headerHash *bbn.BTCHeaderHashBytes) (uint64, error)

	// MainChainDepth returns the depth of the header in the main chain or error if the header does not exist
	MainChainDepth(ctx context.Context, headerBytes *bbn.BTCHeaderHashBytes) (uint64, error)
}

type CheckpointingKeeper interface {
	VerifyCheckpoint(ctx context.Context, checkpoint txformat.RawBtcCheckpoint) error
	// It quite mouthfull to have 4 different methods to operate on checkpoint state
	// but this approach decouples both modules a bit more than having some kind
	// of shared enum passed into the methods. Both modules are free to evolve their
	// representation of checkpoint state independently

	// SetCheckpointSubmitted informs checkpointing module that checkpoint was
	// successfully submitted on btc chain.
	SetCheckpointSubmitted(ctx context.Context, epoch uint64)
	// SetCheckpointConfirmed informs checkpointing module that checkpoint was
	// successfully submitted on btc chain, and it is at least K-deep on the main chain
	SetCheckpointConfirmed(ctx context.Context, epoch uint64)
	// SetCheckpointFinalized informs checkpointing module that checkpoint was
	// successfully submitted on btc chain, and it is at least W-deep on the main chain
	SetCheckpointFinalized(ctx context.Context, epoch uint64)

	// SetCheckpointForgotten informs checkpointing module that this checkpoint lost
	// all submissions on btc chain
	SetCheckpointForgotten(ctx context.Context, epoch uint64)
}

type IncentiveKeeper interface {
	RewardBTCTimestamping(ctx context.Context, epoch uint64, rewardDistInfo *RewardDistInfo)
}
