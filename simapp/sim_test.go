package simapp

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	simappparams "github.com/babylonchain/babylon/app/params"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/babylonchain/babylon/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	btccheckpointtypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
)

// Get flags every time the simulator is run
func init() {
	GetSimulatorFlags()
}

type StoreKeysPrefixes struct {
	A        storetypes.StoreKey
	B        storetypes.StoreKey
	Prefixes [][]byte
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestFullAppSimulation(t *testing.T) {
	config, db, dir, logger, skip, err := SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	privSigner, err := app.SetupPrivSigner()
	require.NoError(t, err)
	babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, FlagPeriodValue, app.GetEncodingConfig(), privSigner, sims.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "BabylonApp", babylon.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		babylon.BaseApp,
		AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(babylon, babylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = sims.CheckExportSimulation(babylon, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		sims.PrintStats(db)
	}
}

func TestAppImportExport(t *testing.T) {
	config, db, dir, logger, skip, err := SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	privSigner, err := app.SetupPrivSigner()
	require.NoError(t, err)
	babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, FlagPeriodValue, app.GetEncodingConfig(), privSigner, sims.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "BabylonApp", babylon.Name())

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		babylon.BaseApp,
		AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(babylon, babylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = sims.CheckExportSimulation(babylon, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		sims.PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := babylon.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newBabylon := app.NewBabylonApp(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, app.DefaultNodeHome, FlagPeriodValue, app.GetEncodingConfig(), privSigner, sims.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "BabylonApp", newBabylon.Name())

	var genesisState app.GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := babylon.NewContext(true, tmproto.Header{Height: babylon.LastBlockHeight()})
	ctxB := newBabylon.NewContext(true, tmproto.Header{Height: babylon.LastBlockHeight()})
	newBabylon.ModuleManager().InitGenesis(ctxB, babylon.AppCodec(), genesisState)
	newBabylon.StoreConsensusParams(ctxB, exported.ConsensusParams)

	fmt.Printf("comparing stores...\n")

	storeKeysPrefixes := []StoreKeysPrefixes{
		{babylon.GetKey(authtypes.StoreKey), newBabylon.GetKey(authtypes.StoreKey), [][]byte{}},
		{babylon.GetKey(stakingtypes.StoreKey), newBabylon.GetKey(stakingtypes.StoreKey),
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey,
			}}, // ordering may change but it doesn't matter
		{babylon.GetKey(slashingtypes.StoreKey), newBabylon.GetKey(slashingtypes.StoreKey), [][]byte{}},
		{babylon.GetKey(minttypes.StoreKey), newBabylon.GetKey(minttypes.StoreKey), [][]byte{}},
		{babylon.GetKey(distrtypes.StoreKey), newBabylon.GetKey(distrtypes.StoreKey), [][]byte{}},
		{babylon.GetKey(banktypes.StoreKey), newBabylon.GetKey(banktypes.StoreKey), [][]byte{banktypes.BalancesPrefix}},
		{babylon.GetKey(paramtypes.StoreKey), newBabylon.GetKey(paramtypes.StoreKey), [][]byte{}},
		{babylon.GetKey(govtypes.StoreKey), newBabylon.GetKey(govtypes.StoreKey), [][]byte{}},
		{babylon.GetKey(evidencetypes.StoreKey), newBabylon.GetKey(evidencetypes.StoreKey), [][]byte{}},
		{babylon.GetKey(capabilitytypes.StoreKey), newBabylon.GetKey(capabilitytypes.StoreKey), [][]byte{}},
		{babylon.GetKey(authzkeeper.StoreKey), newBabylon.GetKey(authzkeeper.StoreKey), [][]byte{}},
		// TODO: add Babylon module StoreKey and prefix here
		{babylon.GetKey(btccheckpointtypes.StoreKey), newBabylon.GetKey(btccheckpointtypes.StoreKey), [][]byte{}},
		{babylon.GetKey(btclightclienttypes.StoreKey), newBabylon.GetKey(btclightclienttypes.StoreKey), [][]byte{}},
		{babylon.GetKey(checkpointingtypes.StoreKey), newBabylon.GetKey(checkpointingtypes.StoreKey), [][]byte{}},
		{babylon.GetKey(epochingtypes.StoreKey), newBabylon.GetKey(epochingtypes.StoreKey),
			[][]byte{epochingtypes.SlashedVotingPowerKey, epochingtypes.VotingPowerKey}},
	}

	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Equal(t, len(failedKVAs), 0, sims.GetSimulationLog(skp.A.Name(), babylon.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

func TestAppSimulationAfterImport(t *testing.T) {
	config, db, dir, logger, skip, err := SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation after import")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	privSigner, err := app.SetupPrivSigner()
	require.NoError(t, err)
	babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, FlagPeriodValue, app.GetEncodingConfig(), privSigner, sims.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "BabylonApp", babylon.Name())

	// Run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		babylon.BaseApp,
		AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(babylon, babylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = sims.CheckExportSimulation(babylon, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		sims.PrintStats(db)
	}

	if stopEarly {
		fmt.Println("can't export or import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := babylon.ExportAppStateAndValidators(true, []string{}, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newBabylon := app.NewBabylonApp(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, app.DefaultNodeHome, FlagPeriodValue, app.GetEncodingConfig(), privSigner, sims.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "BabylonApp", newBabylon.Name())

	newBabylon.InitChain(abci.RequestInitChain{
		AppStateBytes: exported.AppState,
	})

	_, _, err = simulation.SimulateFromSeed(
		t,
		os.Stdout,
		newBabylon.BaseApp,
		AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(newBabylon, newBabylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)
	require.NoError(t, err)
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	if !FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = simappparams.SimAppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			privSigner, err := app.SetupPrivSigner()
			require.NoError(t, err)
			babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, FlagPeriodValue, app.GetEncodingConfig(), privSigner, sims.EmptyAppOptions{}, interBlockCacheOpt())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err = simulation.SimulateFromSeed(
				t,
				os.Stdout,
				babylon.BaseApp,
				AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				SimulationOperations(babylon, babylon.AppCodec(), config),
				babylon.ModuleAccountAddrs(),
				config,
				babylon.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				sims.PrintStats(db)
			}

			appHash := babylon.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
