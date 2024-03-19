package types

const (
	// ModuleName defines the module name
	ModuleName = "btcstaking"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_btcstaking"
)

var (
	ParamsKey               = []byte{0x01} // key prefix for the parameters
	FinalityProviderKey     = []byte{0x02} // key prefix for the finality providers
	BTCDelegatorKey         = []byte{0x03} // key prefix for the BTC delegators
	BTCDelegationKey        = []byte{0x04} // key prefix for the BTC delegations
	VotingPowerKey          = []byte{0x05} // key prefix for the voting power
	BTCHeightKey            = []byte{0x06} // key prefix for the BTC heights
	VotingPowerDistCacheKey = []byte{0x07} // key prefix for voting power distribution cache
	PowerDistUpdateKey      = []byte{0x08} // key prefix for power distribution update events
)
