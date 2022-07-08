package types

const (
	// ModuleName defines the module name
	ModuleName = "epoching"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for epoching
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_epoching"
)

var (
	EpochNumberKey         = []byte{0x11} // key prefix for the epoch number
	QueueLengthKey         = []byte{0x12} // key prefix for the queue length
	QueuedMsgKey           = []byte{0x13} // key prefix for a queued message
	ValidatorSetKey        = []byte{0x14} // key prefix for the validator set in a single epoch
	VotingPowerKey         = []byte{0x15} // key prefix for the total voting power of a validator set in a single epoch
	SlashedVotingPowerKey  = []byte{0x16} // key prefix for the total slashed voting power in a single epoch
	SlashedValidatorSetKey = []byte{0x17} // key prefix for slashed validator set
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
