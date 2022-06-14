package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
	HeadersObjectPrefix = append(HeadersPrefix, 0x0) // where we save the concrete header bytes
	HashToHeightPrefix  = append(HeadersPrefix, 0x1) // where we map hash to height

	TipPrefix = []byte{0x1} // reserve this namespace for the tip
)

func HeadersObjectKey(height uint64, hash []byte) []byte {
	he := sdk.Uint64ToBigEndian(height)

	heightPrefix := append(HeadersObjectPrefix, he...)
	return append(heightPrefix, hash...)
}

func HeadersObjectHeightKey(hash []byte) []byte {
	return append(HashToHeightPrefix, hash...)
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
