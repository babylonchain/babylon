package simapp

import (
	"io/ioutil"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
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
	config.ChainID = helpers.SimAppChainID

	var logger log.Logger
	if FlagVerboseValue {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}

	dir, err := ioutil.TempDir("", dirPrefix)
	if err != nil {
		return simtypes.Config{}, nil, "", nil, false, err
	}

	db, err := sdk.NewLevelDB(dbName, dir)
	if err != nil {
		return simtypes.Config{}, nil, "", nil, false, err
	}

	return config, db, dir, logger, false, nil
}
