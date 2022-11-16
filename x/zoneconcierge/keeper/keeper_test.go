package keeper_test

import (
	"encoding/json"
	"testing"

	"github.com/babylonchain/babylon/app"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
)

// SetupTest creates a coordinator with 2 test chains, and a ZoneConcierge keeper.
func SetupTest(t *testing.T) (*ibctesting.Coordinator, *ibctesting.TestChain, *ibctesting.TestChain, zckeeper.Keeper) {
	var zcKeeper zckeeper.Keeper
	coordinator := ibctesting.NewCoordinator(t, 2)
	// replace the first test chain with a Babylon chain
	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		babylonApp := app.Setup(t, false)
		zcKeeper = babylonApp.ZoneConciergeKeeper
		encCdc := app.MakeTestEncodingConfig()
		genesis := app.NewDefaultGenesisState(encCdc.Marshaler)
		return babylonApp, genesis
	}
	babylonChainID := ibctesting.GetChainID(1)
	coordinator.Chains[babylonChainID] = ibctesting.NewTestChain(t, coordinator, babylonChainID)

	babylonChain := coordinator.GetChain(ibctesting.GetChainID(1))
	czChain := coordinator.GetChain(ibctesting.GetChainID(2))

	return coordinator, babylonChain, czChain, zcKeeper
}
