package simapp

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

    "github.com/babylonchain/babylon/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdksimapp "github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Get flags every time the simulator is run
func init() {
	GetSimulatorFlags()
}

type StoreKeysPrefixes struct {
	A        sdk.StoreKey
	B        sdk.StoreKey
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
	config, db, dir, logger, skip, err := sdksimapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, sdksimapp.FlagPeriodValue, app.MakeTestEncodingConfig(), sdksimapp.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "BabylonApp", babylon.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		babylon.BaseApp,
		AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		sdksimapp.SimulationOperations(babylon, babylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = sdksimapp.CheckExportSimulation(babylon, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		sdksimapp.PrintStats(db)
	}
}

/*
func TestAppImportExport(t *testing.T) {
	config, db, dir, logger, skip, err := sdksimapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, sdksimapp.FlagPeriodValue, app.MakeTestEncodingConfig(), sdksimapp.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "BabylonApp", babylon.Name())

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		babylon.BaseApp,
		AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		sdksimapp.SimulationOperations(babylon, babylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = sdksimapp.CheckExportSimulation(babylon, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		sdksimapp.PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := babylon.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := sdksimapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newBabylon := app.NewBabylonApp(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, app.DefaultNodeHome, sdksimapp.FlagPeriodValue, app.MakeTestEncodingConfig(), sdksimapp.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "BabylonApp", newBabylon.Name())

	var genesisState app.GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := babylon.NewContext(true, tmproto.Header{Height: babylon.LastBlockHeight()})
	ctxB := newBabylon.NewContext(true, tmproto.Header{Height: babylon.LastBlockHeight()})
	newBabylon.mm.InitGenesis(ctxB, babylon.AppCodec(), genesisState)
	newBabylon.StoreConsensusParams(ctxB, exported.ConsensusParams)

	fmt.Printf("comparing stores...\n")

	storeKeysPrefixes := []StoreKeysPrefixes{
		{babylon.keys[authtypes.StoreKey], newBabylon.keys[authtypes.StoreKey], [][]byte{}},
		{babylon.keys[stakingtypes.StoreKey], newBabylon.keys[stakingtypes.StoreKey],
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey,
			}}, // ordering may change but it doesn't matter
		{babylon.keys[slashingtypes.StoreKey], newBabylon.keys[slashingtypes.StoreKey], [][]byte{}},
		{babylon.keys[minttypes.StoreKey], newBabylon.keys[minttypes.StoreKey], [][]byte{}},
		{babylon.keys[distrtypes.StoreKey], newBabylon.keys[distrtypes.StoreKey], [][]byte{}},
		{babylon.keys[banktypes.StoreKey], newBabylon.keys[banktypes.StoreKey], [][]byte{banktypes.BalancesPrefix}},
		{babylon.keys[paramtypes.StoreKey], newBabylon.keys[paramtypes.StoreKey], [][]byte{}},
		{babylon.keys[govtypes.StoreKey], newBabylon.keys[govtypes.StoreKey], [][]byte{}},
		{babylon.keys[evidencetypes.StoreKey], newBabylon.keys[evidencetypes.StoreKey], [][]byte{}},
		{babylon.keys[capabilitytypes.StoreKey], newBabylon.keys[capabilitytypes.StoreKey], [][]byte{}},
		{babylon.keys[authzkeeper.StoreKey], newBabylon.keys[authzkeeper.StoreKey], [][]byte{}},
	}

	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Equal(t, len(failedKVAs), 0, sdksimapp.GetSimulationLog(skp.A.Name(), babylon.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

func TestAppSimulationAfterImport(t *testing.T) {
	config, db, dir, logger, skip, err := sdksimapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation after import")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, sdksimapp.FlagPeriodValue, app.MakeTestEncodingConfig(), sdksimapp.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "BabylonApp", babylon.Name())

	// Run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		babylon.BaseApp,
		AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		sdksimapp.SimulationOperations(babylon, babylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = sdksimapp.CheckExportSimulation(babylon, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		sdksimapp.PrintStats(db)
	}

	if stopEarly {
		fmt.Println("can't export or import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := babylon.ExportAppStateAndValidators(true, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := sdksimapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newBabylon := NewBabylonApp(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, app.DefaultNodeHome, sdksimapp.FlagPeriodValue, app.MakeTestEncodingConfig(), sdksimapp.EmptyAppOptions{}, fauxMerkleModeOpt)
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
		sdksimapp.SimulationOperations(newBabylon, newBabylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)
	require.NoError(t, err)
}
*/

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	if !sdksimapp.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := sdksimapp.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = helpers.SimAppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if sdksimapp.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, sdksimapp.FlagPeriodValue, app.MakeTestEncodingConfig(), sdksimapp.EmptyAppOptions{}, interBlockCacheOpt())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				babylon.BaseApp,
				AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				sdksimapp.SimulationOperations(babylon, babylon.AppCodec(), config),
				babylon.ModuleAccountAddrs(),
				config,
				babylon.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				sdksimapp.PrintStats(db)
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
