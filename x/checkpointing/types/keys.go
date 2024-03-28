package types

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
	CheckpointsPrefix        = []byte{0x1} // reserve this namespace for checkpoints
	RegistrationPrefix       = []byte{0x2} // reserve this namespace for BLS keys
	ValidatorBlsKeySetPrefix = []byte{0x3} // reserve this namespace for validator BLS key set

	CkptsObjectPrefix = append(CheckpointsPrefix, 0x0) // where we save the concrete BLS sig bytes

	AddrToBlsKeyPrefix = append(RegistrationPrefix, 0x0) // where we save the concrete BLS public keys
	BlsKeyToAddrPrefix = append(RegistrationPrefix, 0x1) // where we save BLS key set

	FinalizedEpochKey = []byte{0x04} // FinalizedEpochKey defines the key to store the last finalised epoch
)

// CkptsObjectKey defines epoch
func CkptsObjectKey(epoch uint64) []byte {
	return sdk.Uint64ToBigEndian(epoch)
}

// ValidatorBlsKeySetKey defines epoch
func ValidatorBlsKeySetKey(epoch uint64) []byte {
	return sdk.Uint64ToBigEndian(epoch)
}

// AddrToBlsKeyKey defines validator address
func AddrToBlsKeyKey(valAddr sdk.ValAddress) []byte {
	return valAddr
}

// BlsKeyToAddrKey defines BLS public key
func BlsKeyToAddrKey(pk bls12381.PublicKey) []byte {
	return pk
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
