package types

const (
	// ModuleName defines the module name
	ModuleName = "finality"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_finality"
)

var (
	BlockKey                = []byte{0x01} // key prefix for blocks
	VoteKey                 = []byte{0x02} // key prefix for votes
	PubRandKey              = []byte{0x03} // key prefix for public randomness
	ParamsKey               = []byte{0x04} // key prefix for the parameters
	EvidenceKey             = []byte{0x05} // key prefix for evidences
	NextHeightToFinalizeKey = []byte{0x06} // key prefix for next height to finalise
)
