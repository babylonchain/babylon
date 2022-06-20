package types

import (
	"github.com/btcsuite/btcd/wire"
)

// TODO Mock keepers are currently only used when wiring app to satisfy the compiler
type MockBTCLightClientKeeper struct{}
type MockCheckpointingKeeper struct{}

func (mb MockBTCLightClientKeeper) BlockHeight(header wire.BlockHeader) (uint64, error) {
	return uint64(10), nil
}

func (ck MockCheckpointingKeeper) CheckpointValid(rawCheckpoint []byte) (uint64, error) {
	return uint64(10), nil
}
