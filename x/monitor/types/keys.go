package types

import (
	"fmt"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "monitor"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_monitor"
)

var (
	EpochEndLightClientHeightPrefix           = []byte{1}
	CheckpointReportedLightClientHeightPrefix = []byte{2}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func GetEpochEndLightClientHeightKey(e uint64) []byte {
	return append(EpochEndLightClientHeightPrefix, sdk.Uint64ToBigEndian(e)...)
}

func GetCheckpointReportedLightClientHeightKey(hashString string) ([]byte, error) {
	hashBytes, err := types.FromStringToCkptHash(hashString)
	if err != nil {
		return nil, fmt.Errorf("invalid hash string %s: %w", hashString, err)
	}
	return append(CheckpointReportedLightClientHeightPrefix, hashBytes...), nil
}
