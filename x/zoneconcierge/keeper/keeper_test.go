package keeper_test

import (
	"encoding/json"
	"testing"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
)

// SetupTest creates a coordinator with 2 test chains, and a ZoneConcierge keeper.
func SetupTest(t *testing.T) (*ibctesting.Coordinator, *ibctesting.TestChain, *ibctesting.TestChain, *app.BabylonApp) {
	var bbnApp *app.BabylonApp
	coordinator := ibctesting.NewCoordinator(t, 2)
	// replace the first test chain with a Babylon chain
	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		babylonApp := app.Setup(t, false)
		bbnApp = babylonApp
		encCdc := app.MakeTestEncodingConfig()
		genesis := app.NewDefaultGenesisState(encCdc.Marshaler)
		return babylonApp, genesis
	}
	babylonChainID := ibctesting.GetChainID(1)
	coordinator.Chains[babylonChainID] = ibctesting.NewTestChain(t, coordinator, babylonChainID)

	babylonChain := coordinator.GetChain(ibctesting.GetChainID(1))
	czChain := coordinator.GetChain(ibctesting.GetChainID(2))

	return coordinator, babylonChain, czChain, bbnApp
}

// SimulateHeadersViaHook generates a non-zero number of canonical headers via the hook
func SimulateHeadersViaHook(ctx sdk.Context, hooks zckeeper.Hooks, chainID string, numHeaders uint64) []*ibctmtypes.Header {
	headers := []*ibctmtypes.Header{}
	// invoke the hook a number of times to simulate a number of blocks
	for i := uint64(0); i < numHeaders; i++ {
		header := datagen.GenRandomIBCTMHeader(chainID, i)
		hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
		headers = append(headers, header)
	}
	return headers
}

// SimulateHeadersViaHook generates a random non-zero number of canonical headers and fork headers via the hook
func SimulateHeadersAndForksViaHook(ctx sdk.Context, hooks zckeeper.Hooks, chainID string, numHeaders uint64, numForkHeaders uint64) ([]*ibctmtypes.Header, []*ibctmtypes.Header) {
	headers := []*ibctmtypes.Header{}
	// invoke the hook a number of times to simulate a number of blocks
	for i := uint64(0); i < numHeaders; i++ {
		header := datagen.GenRandomIBCTMHeader(chainID, i)
		hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
		headers = append(headers, header)
	}

	// generate a number of fork headers
	forkHeaders := []*ibctmtypes.Header{}
	for i := uint64(0); i < numForkHeaders; i++ {
		header := datagen.GenRandomIBCTMHeader(chainID, numHeaders-1)
		hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, true)
		forkHeaders = append(forkHeaders, header)
	}
	return headers, forkHeaders
}
