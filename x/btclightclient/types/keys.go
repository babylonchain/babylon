package types

import (
	bbn "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "btclightclient"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_btclightclient"
)

var (
	HeadersPrefix       = []byte{0x0}                // reserve this namespace for headers
	HeadersObjectPrefix = append(HeadersPrefix, 0x0) // reserve this namespace mapping: Height -> BTCHeaderInfo
	HashToHeightPrefix  = append(HeadersPrefix, 0x1) // reserve this namespace mapping: Hash -> Height
)

func HeadersObjectKey(height uint64) []byte {
	return sdk.Uint64ToBigEndian(height)
}

func HeadersObjectHeightKey(hash *bbn.BTCHeaderHashBytes) []byte {
	return hash.MustMarshal()
}
