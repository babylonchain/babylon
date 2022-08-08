package types

import (
	bbl "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO Mock keepers are currently only used when wiring app to satisfy the compiler
type MockBTCLightClientKeeper struct{}
type MockCheckpointingKeeper struct{}

func (mb MockBTCLightClientKeeper) BlockHeight(ctx sdk.Context, header *bbl.BTCHeaderHashBytes) (uint64, error) {
	return uint64(10), nil
}

func (mb MockBTCLightClientKeeper) IsAncestor(ctx sdk.Context, parentHash *bbl.BTCHeaderHashBytes, childHash *bbl.BTCHeaderHashBytes) (bool, error) {
	return true, nil
}

func (ck MockCheckpointingKeeper) CheckpointEpoch(ctx sdk.Context, rawCheckpoint []byte) (uint64, error) {
	return uint64(10), nil
}

// SetCheckpointSubmitted Informs checkpointing module that checkpoint was
// successfully submitted on btc chain.
func (ck MockCheckpointingKeeper) SetCheckpointSubmitted(ctx sdk.Context, epoch uint64) {
}

// SetCheckpointConfirmed Informs checkpointing module that checkpoint was
// successfully submitted on btc chain, and it is at least K-deep on the main chain
func (ck MockCheckpointingKeeper) SetCheckpointConfirmed(ctx sdk.Context, epoch uint64) {
}

// SetCheckpointFinalized Informs checkpointing module that checkpoint was
// successfully submitted on btc chain, and it is at least W-deep on the main chain
func (ck MockCheckpointingKeeper) SetCheckpointFinalized(ctx sdk.Context, epoch uint64) {
}

// SetCheckpointForgotten Informs checkpointing module that was in submitted state
// lost all its checkpoints and is checkpoint empty
func (ck MockCheckpointingKeeper) SetCheckpointForgotten(ctx sdk.Context, epoch uint64) {
}

func (ck MockBTCLightClientKeeper) MainChainDepth(ctx sdk.Context, headerBytes *bbl.BTCHeaderHashBytes) (int64, error) {
	return 1, nil
}
