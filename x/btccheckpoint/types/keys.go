package types

import "github.com/cosmos/cosmos-sdk/codec"

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
	UnconfirmedIndexPrefix = []byte{0, 0, 0, 0, 0, 0}
	ConfirmedIndexPrefix   = []byte{1, 1, 1, 1, 1, 1}
	FinalizedIndexPrefix   = []byte{2, 2, 2, 2, 2, 2}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
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
