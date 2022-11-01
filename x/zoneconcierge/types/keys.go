package types

const (
	// ModuleName defines the module name
	ModuleName = "zoneconcierge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_zoneconcierge"

	// Version defines the current version the IBC module supports
	Version = "zoneconcierge-1"

	// PortID is the default port id that module binds to
	PortID = "zoneconcierge"
)

var (
	PortKey           = []byte{0x11} // PortKey defines the key to store the port ID in store
	ChainInfoKey      = []byte{0x12} // ChainInfoKey defines the key to store the chain info for each CZ in store
	CanonicalChainKey = []byte{0x13} // CanonicalChainKey defines the key to store the canonical chain for each CZ in store
	ForkKey           = []byte{0x14} // ForkKey defines the key to store the forks for each CZ in store
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
