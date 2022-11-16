package keeper_test

import (
	"encoding/json"
	"testing"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func SimulateHeadersViaHook(ctx sdk.Context, hooks zckeeper.Hooks, chainID string) uint64 {
	// invoke the hook a random number of times to simulate a random number of blocks
	numHeaders := datagen.RandomInt(100) + 1
	for i := uint64(0); i < numHeaders; i++ {
		header := datagen.GenRandomIBCTMHeader(chainID, i)
		hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
	}
	return numHeaders
}

func SimulateHeadersAndForksViaHook(ctx sdk.Context, hooks zckeeper.Hooks, chainID string) (uint64, uint64) {
	// invoke the hook a random number of times to simulate a random number of blocks
	numHeaders := datagen.RandomInt(100) + 1
	for i := uint64(0); i < numHeaders; i++ {
		header := datagen.GenRandomIBCTMHeader(chainID, i)
		hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
	}

	// generate a number of fork headers
	numForkHeaders := datagen.RandomInt(10) + 1
	for i := uint64(0); i < numForkHeaders; i++ {
		header := datagen.GenRandomIBCTMHeader(chainID, numHeaders-1)
		hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, true)
	}
	return numHeaders, numForkHeaders
}
