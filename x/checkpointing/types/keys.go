package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "checkpointing"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_checkpointing"
)

var (
	BlsSigsPrefix     = []byte{0x0} // reserve this namespace for bls sigs
	CheckpointsPrefix = []byte{0x1} // reserve this namespace for checkpoints

	BlsSigsObjectPrefix      = append(BlsSigsPrefix, 0x0) // where we save the concrete bls sig bytes
	BlsSigsHashToEpochPrefix = append(BlsSigsPrefix, 0x1) // where we map hash to epoch

	CkptsObjectPrefix       = append(CheckpointsPrefix, 0x0) // where we save the concrete bls sig bytes
	CkptsHashToEpochPrefix  = append(CheckpointsPrefix, 0x1) // where we map hash to epoch
	CkptsHashToStatusPrefix = append(CheckpointsPrefix, 0x2) // where we map hash to status
	TipPrefix               = append(CheckpointsPrefix, 0x3) // where we store the tip epoch
	LastConfirmedPrefix     = append(CheckpointsPrefix, 0x4) // where we store the last confirmed epoch
)

// BlsSigsObjectKey defines epoch + hash
func BlsSigsObjectKey(epoch uint64, hash BlsSigHash) []byte {
	ee := sdk.Uint64ToBigEndian(epoch)
	epochPrefix := append(BlsSigsObjectPrefix, ee...)
	return append(epochPrefix, hash...)
}

func BlsSigsEpochKey(hash BlsSigHash) []byte {
	return append(BlsSigsHashToEpochPrefix, hash...)
}

// CkptsObjectKey defines epoch + status + hash
func CkptsObjectKey(epoch uint64, status uint32, hash RawCkptHash) []byte {
	ee := sdk.Uint64ToBigEndian(epoch)
	epochPrefix := append(CkptsObjectPrefix, ee...)
	epochStatusPrefix := append(epochPrefix, Uint32ToBitEndian(status)...)
	return append(epochStatusPrefix, hash...)
}

func CkptsEpochKey(hash RawCkptHash) []byte {
	return append(CkptsHashToEpochPrefix, hash...)
}

func CkptsStatusKey(hash RawCkptHash) []byte {
	return append(CkptsHashToStatusPrefix, hash...)
}

func TipKey() []byte {
	return TipPrefix
}

func LastConfirmedKey() []byte {
	return LastConfirmedPrefix
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
