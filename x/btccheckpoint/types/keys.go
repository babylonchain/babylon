package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "btccheckpoint"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	TStoreKey = "transient_btccheckpoint"

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_btccheckpoint"

	LatestFinalizedEpochKey = "latestFinalizedEpoch"

	btcLightClientUpdated = "btcLightClientUpdated"
)

var (
	SubmisionKeyPrefix       = []byte{3}
	EpochDataPrefix          = []byte{4}
	LastFinalizedEpochKey    = append([]byte{5}, []byte(LatestFinalizedEpochKey)...)
	BtcLightClientUpdatedKey = append([]byte{6}, []byte(btcLightClientUpdated)...)
	ParamsKey                = []byte{7}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func PrefixedSubmisionKey(cdc codec.BinaryCodec, k *SubmissionKey) []byte {
	return append(SubmisionKeyPrefix, cdc.MustMarshal(k)...)
}

func GetEpochIndexKey(e uint64) []byte {
	return append(EpochDataPrefix, sdk.Uint64ToBigEndian(e)...)
}

func GetLatestFinalizedEpochKey() []byte {
	return LastFinalizedEpochKey
}

func GetBtcLightClientUpdatedKey() []byte {
	return BtcLightClientUpdatedKey
}
