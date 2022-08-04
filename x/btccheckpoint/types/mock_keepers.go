package types

import (
	"errors"

	bbl "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MockBTCLightClientKeeper struct {
	depth       int64
	returnError bool
}

type MockCheckpointingKeeper struct {
	epoch       uint64
	returnError bool
}

func NewMockBTCLightClientKeeper(initialDepth int64) *MockBTCLightClientKeeper {
	lc := MockBTCLightClientKeeper{
		depth:       initialDepth,
		returnError: false,
	}
	return &lc
}

func NewMockCheckpointingKeeper(epoch uint64) *MockCheckpointingKeeper {
	mc := MockCheckpointingKeeper{
		epoch:       epoch,
		returnError: false,
	}
	return &mc
}

func (mc *MockCheckpointingKeeper) SetEpoch(e uint64) {
	mc.epoch = e
}

func (mc *MockCheckpointingKeeper) ReturnError() {
	mc.returnError = true
}

func (mc *MockCheckpointingKeeper) ReturnSuccess() {
	mc.returnError = false
}

func (mc *MockBTCLightClientKeeper) SetDepth(d int64) {
	mc.depth = d
}

func (mc *MockBTCLightClientKeeper) ReturnError() {
	mc.returnError = true
}

func (mc *MockBTCLightClientKeeper) ReturnSuccess() {
	mc.returnError = false
}

func (mb MockBTCLightClientKeeper) BlockHeight(ctx sdk.Context, header *bbl.BTCHeaderHashBytes) (uint64, error) {
	// todo not used
	return uint64(10), nil
}

func (mb MockBTCLightClientKeeper) IsAncestor(ctx sdk.Context, parentHash *bbl.BTCHeaderHashBytes, childHash *bbl.BTCHeaderHashBytes) (bool, error) {
	return true, nil
}

func (ck MockBTCLightClientKeeper) MainChainDepth(ctx sdk.Context, headerBytes *bbl.BTCHeaderHashBytes) (int64, error) {
	if ck.returnError {
		return -1, errors.New("unknown header")
	}

	return ck.depth, nil
}

func (ck MockCheckpointingKeeper) CheckpointEpoch(rawCheckpoint []byte) (uint64, error) {
	if ck.returnError {
		return 0, errors.New("bad checkpoints")
	}

	return ck.epoch, nil
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
