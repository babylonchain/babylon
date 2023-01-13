package types

import (
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
	EpochEndLightClientHeightPrefix = []byte{1}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func GetEpochEndLightClientHeightKey(e uint64) []byte {
	return append(EpochEndLightClientHeightPrefix, sdk.Uint64ToBigEndian(e)...)
}
