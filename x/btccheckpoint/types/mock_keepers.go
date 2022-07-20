package types

import (
	btypes "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO Mock keepers are currently only used when wiring app to satisfy the compiler
type MockBTCLightClientKeeper struct{}
type MockCheckpointingKeeper struct{}

func (mb MockBTCLightClientKeeper) BlockHeight(ctx sdk.Context, header btypes.BTCHeaderHashBytes) (uint64, error) {
	return uint64(10), nil
}

func (mb MockBTCLightClientKeeper) IsAncestor(ctx sdk.Context, parentHash btypes.BTCHeaderHashBytes, childHash btypes.BTCHeaderHashBytes) (bool, error) {
	return true, nil
}

func (ck MockCheckpointingKeeper) CheckpointEpoch(rawCheckpoint []byte) (uint64, error) {
	return uint64(10), nil
}

// SetCheckpointSubmitted Informs checkpointing module that checkpoint was
// sucessfully submitted on btc chain. It can be either or main chain or fork.
func (ck MockCheckpointingKeeper) SetCheckpointSubmitted(rawCheckpoint []byte) {}

// SetCheckpointSubmitted Informs checkpointing module that checkpoint was
// sucessfully submitted on btc chain and it is at least K-deep on the main chain
func (ck MockCheckpointingKeeper) SetCheckpointConfirmed(rawCheckpoint []byte) {}

// SetCheckpointSubmitted Informs checkpointing module that checkpoint was
// sucessfully submitted on btc chain and it is at least W-deep on the main chain
func (ck MockCheckpointingKeeper) SetCheckpointFinalized(rawCheckpoint []byte) {}

// SetCheckpointForgotten Informs checkpointing module that was in submitted state
// lost all its checkpoints and is checkpoint empty
func (ck MockCheckpointingKeeper) SetCheckpointForgotten(rawCheckpoint []byte) {}

func (ck MockBTCLightClientKeeper) ChainDepth(ctx sdk.Context, headerBytes *btypes.BTCHeaderHashBytes) (uint64, bool, error) {
	return 1, true, nil
}
