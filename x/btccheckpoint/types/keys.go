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

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_btccheckpoint"
)

var (
	SubmisionKeyPrefix     = []byte{3}
	UnconfirmedIndexPrefix = []byte{4}
	ConfirmedIndexPrefix   = []byte{5}
	FinalizedIndexPrefix   = []byte{6}
	EpochDataPrefix        = []byte{7}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func PrefixedSubmisionKey(cdc codec.BinaryCodec, k *SubmissionKey) []byte {
	return append(SubmisionKeyPrefix, cdc.MustMarshal(k)...)
}

func UnconfiredSubmissionsKey(cdc codec.BinaryCodec, k *SubmissionKey) []byte {
	return append(UnconfirmedIndexPrefix, cdc.MustMarshal(k)...)
}

func ConfirmedSubmissionsKey(cdc codec.BinaryCodec, k *SubmissionKey) []byte {
	return append(ConfirmedIndexPrefix, cdc.MustMarshal(k)...)
}

func FinalizedSubmissionsKey(cdc codec.BinaryCodec, k *SubmissionKey) []byte {
	return append(FinalizedIndexPrefix, cdc.MustMarshal(k)...)
}

func GetEpochIndexKey(e uint64) []byte {
	return append(EpochDataPrefix, sdk.Uint64ToBigEndian(e)...)
}
