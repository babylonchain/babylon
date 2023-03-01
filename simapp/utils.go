package simapp

import (
	"encoding/json"
	"os"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	simappparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/app"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// SetupSimulation creates the config, db (levelDB), temporary directory and logger for the simulation tests.
// If `FlagEnabledValue` is false it skips the current test.
// Returns error on an invalid db intantiation or temp dir creation.
// NOTE: this function is identical to https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/simapp/utils.go.
// The reason of migrating it here is that it uses the modules' flags initialised in `init()`. Otherwise,
// if using  `sdksimapp.SetupSimulation`, then `sdksimapp.FlagEnabledValue` and `sdksimapp.FlagVerboseValue`
// will be used, rather than those in `config.go` of our module.  The other functions under
// https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/simapp/utils.go do not access flags and thus can be reused safely.
func SetupSimulation(dirPrefix, dbName string) (simtypes.Config, dbm.DB, string, log.Logger, bool, error) {
	if !FlagEnabledValue {
		return simtypes.Config{}, nil, "", nil, true, nil
	}

	config := NewConfigFromFlags()
	config.ChainID = simappparams.SimAppChainID

	var logger log.Logger
	if FlagVerboseValue {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}

	dir, err := os.MkdirTemp("", dirPrefix)
	if err != nil {
		return simtypes.Config{}, nil, "", nil, false, err
	}

	db, err := dbm.NewDB(dbName, dbm.BackendType(config.DBBackend), dir)
	if err != nil {
		return simtypes.Config{}, nil, "", nil, false, err
	}

	return config, db, dir, logger, false, nil
}

// SimulationOperations retrieves the simulation params from the provided file path
// and returns all the modules weighted operations
// NOTE: the code is same as https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/simapp/utils.go#L50-L73,
// except that this function modifies the default weights given in Cosmos SDK.
// Specifically, Babylon does not want unwrapped message types in the staking module.
func SimulationOperations(app app.App, cdc codec.JSONCodec, config simtypes.Config) []simtypes.WeightedOperation {
	simState := module.SimulationState{
		AppParams: make(simtypes.AppParams),
		Cdc:       cdc,
	}

	if config.ParamsFile != "" {
		bz, err := os.ReadFile(config.ParamsFile)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(bz, &simState.AppParams)
		if err != nil {
			panic(err)
		}
	}
	// get weighted operations from all modules, except for the staking module whose messages will be rejected by Babylon
	sm := app.SimulationManager()
	appWOps := make([]simtypes.WeightedOperation, 0, len(sm.Modules))
	for _, module := range sm.Modules {
		if _, ok := module.(staking.AppModule); ok {
			continue
		}
		appWOps = append(appWOps, module.WeightedOperations(simState)...)
	}

	return appWOps
}
