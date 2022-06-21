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
	EpochNumberKey = []byte{0x11} // key prefix for the epoch number
	QueueLengthKey = []byte{0x12} // key prefix for the queue length
	QueuedMsgKey   = []byte{0x13} // key prefix for a queued message
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
