package types

import (
	"errors"

	bbn "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MockBTCLightClientKeeper struct {
	headers map[string]int64
}

type MockCheckpointingKeeper struct {
	epoch       uint64
	returnError bool
}

func NewMockBTCLightClientKeeper() *MockBTCLightClientKeeper {
	lc := MockBTCLightClientKeeper{
		headers: make(map[string]int64),
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

func (mc *MockBTCLightClientKeeper) SetDepth(header *bbn.BTCHeaderHashBytes, dd int64) {
	mc.headers[header.String()] = dd
}

func (mb MockBTCLightClientKeeper) BlockHeight(ctx sdk.Context, header *bbn.BTCHeaderHashBytes) (uint64, error) {
	// todo not used
	return uint64(10), nil
}

func (ck MockBTCLightClientKeeper) MainChainDepth(ctx sdk.Context, headerBytes *bbn.BTCHeaderHashBytes) (int64, error) {
	depth, ok := ck.headers[headerBytes.String()]
	if ok {
		return depth, nil
	} else {
		return 0, errors.New("unknown header")
	}
}

func (ck MockCheckpointingKeeper) CheckpointEpoch(ctx sdk.Context, rawCheckpoint []byte) (uint64, error) {
	if ck.returnError {
		return 0, errors.New("bad checkpoints")
	}

	return ck.epoch, nil
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
