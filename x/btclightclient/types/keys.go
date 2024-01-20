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
	HeadersObjectPrefix  = []byte{0x01} // reserve this namespace mapping: Height -> BTCHeaderInfo
	HashToHeightPrefix   = []byte{0x02} // reserve this namespace mapping: Hash -> Height
	LastRollbackPointKey = []byte{0x03} // the point of the last reorg
	ParamsKey            = []byte{0x04} // key for params
)

func HeadersObjectKey(height uint64) []byte {
	return sdk.Uint64ToBigEndian(height)
}

func HeadersObjectHeightKey(hash *bbn.BTCHeaderHashBytes) []byte {
	return hash.MustMarshal()
}
