package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// TODO Mock keepers are currently only used when wiring app to satisfy the compiler
type MockBTCLightClientKeeper struct{}
type MockCheckpointingKeeper struct{}

func (mb MockBTCLightClientKeeper) BlockHeight(header chainhash.Hash) (uint64, error) {
	return uint64(10), nil
}

func (mb MockBTCLightClientKeeper) IsAncestor(parentHash chainhash.Hash, childHash chainhash.Hash) (bool, error) {
	return true, nil
}

func (ck MockCheckpointingKeeper) CheckpointEpoch(rawCheckpoint []byte) (uint64, error) {
	return uint64(10), nil
}
