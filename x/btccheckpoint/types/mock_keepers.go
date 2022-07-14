package types

import (
	btypes "github.com/babylonchain/babylon/types"
)

// TODO Mock keepers are currently only used when wiring app to satisfy the compiler
type MockBTCLightClientKeeper struct{}
type MockCheckpointingKeeper struct{}

func (mb MockBTCLightClientKeeper) BlockHeight(header btypes.BTCHeaderHashBytes) (uint64, error) {
	return uint64(10), nil
}

func (mb MockBTCLightClientKeeper) IsAncestor(parentHash btypes.BTCHeaderHashBytes, childHash btypes.BTCHeaderHashBytes) (bool, error) {
	return true, nil
}

func (ck MockCheckpointingKeeper) CheckpointEpoch(rawCheckpoint []byte) (uint64, error) {
	return uint64(10), nil
}
