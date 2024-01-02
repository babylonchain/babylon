package configurer

import (
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"testing"

	"github.com/babylonchain/babylon/test/e2e/configurer/chain"
	"github.com/babylonchain/babylon/test/e2e/containers"
	"github.com/babylonchain/babylon/test/e2e/initialization"
)

type Configurer interface {
	ConfigureChains() error

	ClearResources() error

	GetChainConfig(chainIndex int) *chain.Config

	RunSetup() error

	RunValidators() error

	InstantiateBabylonContract() error

	RunIBC() error
}

var (
	// Last nodes are non validator nodes to serve as the ones using relayer. Out
	// validators are constantly sending bls transactions which make relayer operations
	// fail constantly

	// each started validator container corresponds to one of
	// the configurations below.
	validatorConfigsChainA = []*initialization.NodeConfig{
		{
			// this is a node that is used to state-sync from so its snapshot-interval
			// is frequent.
			Name:               "babylon-default-a-1",
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   25,
			SnapshotKeepRecent: 10,
			IsValidator:        true,
		},
		{
			Name:               "babylon-default-a-2",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "babylon-default-a-3",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        false,
		},
	}
	validatorConfigsChainB = []*initialization.NodeConfig{
		{
			Name:               "babylon-default-b-1",
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "babylon-default-b-2",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "babylon-default-b-3",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        false,
		},
	}
	ibcConfigChainA = &ibctesting.ChannelConfig{
		PortID:  zctypes.PortID,
		Order:   zctypes.Ordering,
		Version: zctypes.Version,
	}
	ibcConfigChainB = &ibctesting.ChannelConfig{
		PortID:  zctypes.PortID, // Will be replaced by the contract address in Phase 2 tests
		Order:   zctypes.Ordering,
		Version: zctypes.Version,
	}
)

// NewBTCTimestampingConfigurer returns a new Configurer for BTC timestamping service.
// TODO currently only one configuration is available. Consider testing upgrades
// when necessary
func NewBTCTimestampingConfigurer(t *testing.T, isDebugLogEnabled bool) (Configurer, error) {
	containerManager, err := containers.NewManager(isDebugLogEnabled)
	if err != nil {
		return nil, err
	}

	return NewCurrentBranchConfigurer(t,
		[]*chain.Config{
			chain.New(t, containerManager, initialization.ChainAID, validatorConfigsChainA, ibcConfigChainA),
			chain.New(t, containerManager, initialization.ChainBID, validatorConfigsChainB, ibcConfigChainB),
		},
		withIBC(baseSetup), // base set up with IBC
		containerManager,
	), nil
}

// NewBTCTimestampingPhase2Configurer returns a new Configurer for BTC timestamping service (phase 2).
func NewBTCTimestampingPhase2Configurer(t *testing.T, isDebugLogEnabled bool) (Configurer, error) {
	containerManager, err := containers.NewManager(isDebugLogEnabled)
	if err != nil {
		return nil, err
	}

	return NewCurrentBranchConfigurer(t,
		[]*chain.Config{
			chain.New(t, containerManager, initialization.ChainAID, validatorConfigsChainA, ibcConfigChainA),
			chain.New(t, containerManager, initialization.ChainBID, validatorConfigsChainB, ibcConfigChainB),
		},
		withPhase2IBC(baseSetup), // IBC setup (requires contract address)
		containerManager,
	), nil
}

// NewBTCStakingConfigurer returns a new Configurer for BTC staking service
func NewBTCStakingConfigurer(t *testing.T, isDebugLogEnabled bool) (Configurer, error) {
	containerManager, err := containers.NewManager(isDebugLogEnabled)
	if err != nil {
		return nil, err
	}

	return NewCurrentBranchConfigurer(t,
		[]*chain.Config{
			// we only need 1 chain for testing BTC staking
			chain.New(t, containerManager, initialization.ChainAID, validatorConfigsChainA, nil),
		},
		baseSetup, // base set up
		containerManager,
	), nil
}
