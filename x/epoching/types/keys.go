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

// QueuedMsgKey is the key prefix for a queued message
var (
	EpochNumberKey = []byte{0x11}
	QueuedMsgKey   = []byte{0x012}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
