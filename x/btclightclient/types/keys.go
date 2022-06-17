package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
	HeadersObjectPrefix = append(HeadersPrefix, 0x0) // where we save the concrete header bytes
	HashToHeightPrefix  = append(HeadersPrefix, 0x1) // where we map hash to height

	TipPrefix = []byte{0x1} // reserve this namespace for the tip
)

func HeadersObjectKey(height uint64, hash *chainhash.Hash) []byte {
	he := sdk.Uint64ToBigEndian(height)

	hashBytes := ChainhashToBytes(hash)

	heightPrefix := append(HeadersObjectPrefix, he...)
	return append(heightPrefix, hashBytes...)
}

func HeadersObjectHeightKey(hash *chainhash.Hash) []byte {
	hashBytes := ChainhashToBytes(hash)
	return append(HashToHeightPrefix, hashBytes...)
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
