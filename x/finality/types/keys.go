package types

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/babylonchain/babylon/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "finality"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_finality"

	// MissedBlockBitmapChunkSize defines the chunk size, in number of bits, of a
	// finality provider missed block bitmap. Chunks are used to reduce the storage and
	// write overhead of IAVL nodes. The total size of the bitmap is roughly in
	// the range [0, SignedBlocksWindow) where each bit represents a block. A
	// finality provider's IndexOffset modulo the SignedBlocksWindow is used to retrieve
	// the chunk in that bitmap range. Once the chunk is retrieved, the same index
	// is used to check or flip a bit, where if a bit is set, it indicates the
	// finality provider missed that block.
	//
	// For a bitmap of N items, i.e. a finality provider's signed block window, the amount
	// of write complexity per write with a factor of f being the overhead of
	// IAVL being un-optimized, i.e. 2-4, is as follows:
	//
	// ChunkSize + (f * 256 <IAVL leaf hash>) + 256 * log_2(N / ChunkSize)
	//
	// As for the storage overhead, with the same factor f, it is as follows:
	// (N - 256) + (N / ChunkSize) * (512 * f)
	MissedBlockBitmapChunkSize = 1024 // 2^10 bits
)

var (
	BlockKey                                   = []byte{0x01}             // key prefix for blocks
	VoteKey                                    = []byte{0x02}             // key prefix for votes
	PubRandKey                                 = []byte{0x03}             // key prefix for public randomness
	PubRandCommitKey                           = []byte{0x04}             // key prefix for commitment of public randomness
	ParamsKey                                  = []byte{0x05}             // key prefix for the parameters
	EvidenceKey                                = []byte{0x06}             // key prefix for evidences
	NextHeightToFinalizeKey                    = []byte{0x07}             // key prefix for next height to finalise
	FinalityProviderSigningInfoKeyPrefix       = collections.NewPrefix(8) // key prefix for signing info
	FinalityProviderMissedBlockBitmapKeyPrefix = collections.NewPrefix(9) // key prefix for missed block bitmap
)

// FinalityProviderSigningInfoKey - stored by finality provider public key in BIP340
func FinalityProviderSigningInfoKey(pk *types.BIP340PubKey) []byte {
	return append(FinalityProviderSigningInfoKeyPrefix, address.MustLengthPrefix(pk.MustMarshal())...)
}

// FinalityProviderMissedBlockBitmapKey - stored by finality provider public key in BIP340
func FinalityProviderMissedBlockBitmapKey(pk *types.BIP340PubKey) []byte {
	return append(FinalityProviderMissedBlockBitmapKeyPrefix, address.MustLengthPrefix(pk.MustMarshal())...)
}
