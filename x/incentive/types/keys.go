package types

const (
	// ModuleName defines the module name
	ModuleName = "incentive"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_incentive"
)

var (
	ParamsKey               = []byte{0x01} // key prefix for the parameters
	BTCStakingGaugeKey      = []byte{0x02} // key prefix for BTC staking gauge at each height
	BTCTimestampingGaugeKey = []byte{0x03} // key prefix for BTC timestamping gauge at each height
	RewardGaugeKey          = []byte{0x04} // key prefix for reward gauge for a given stakeholder in a given type
)
