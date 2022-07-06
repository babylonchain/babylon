package types

import (
	bbl "github.com/babylonchain/babylon/types"
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
	TipPrefix           = append(HeadersPrefix, 0x2) // where we store the tip
)

func HeadersObjectKey(height uint64, hash *chainhash.Hash) ([]byte, error) {
	he := sdk.Uint64ToBigEndian(height)

	hashBytes, err := bbl.NewBTCHeaderHashBytesFromChainhash(hash)
	if err != nil {
		return nil, err
	}

	heightPrefix := append(HeadersObjectPrefix, he...)
	return append(heightPrefix, hashBytes...), nil
}

func HeadersObjectHeightKey(hash *chainhash.Hash) ([]byte, error) {
	hashBytes, err := bbl.NewBTCHeaderHashBytesFromChainhash(hash)
	if err != nil {
		return nil, err
	}
	return append(HashToHeightPrefix, hashBytes...), nil
}

}

func TipKey() []byte {
	return TipPrefix
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
