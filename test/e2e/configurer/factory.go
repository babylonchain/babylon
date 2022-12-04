package configurer

import (
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

	RunIBC() error
}

var (
	// each started validator containers corresponds to one of
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
	}
)

// New returns a new Configurer.
// TODO currently only one configuration is available. Consider testing upgrades
// when necessary
func New(t *testing.T, isDebugLogEnabled bool) (Configurer, error) {
	containerManager, err := containers.NewManager(isDebugLogEnabled)
	if err != nil {
		return nil, err
	}

	return NewCurrentBranchConfigurer(t,
		[]*chain.Config{
			chain.New(t, containerManager, initialization.ChainAID, validatorConfigsChainA),
			chain.New(t, containerManager, initialization.ChainBID, validatorConfigsChainB),
		},
		withIBC(baseSetup), // base set up with IBC
		containerManager,
	), nil
}
