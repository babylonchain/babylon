package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Simulation parameter constants
const (
	EpochIntervalKey = "epoch_interval"
)

// genEpochInterval returns randomized EpochInterval
func genEpochInterval(r *rand.Rand) uint64 {
	return uint64(r.Intn(10) + 2)
}

// RandomizedGenState generates a random GenesisState for staking
func RandomizedGenState(simState *module.SimulationState) {
	var epochInterval uint64
	simState.AppParams.GetOrGenerate(
		EpochIntervalKey, &epochInterval, simState.Rand,
		func(r *rand.Rand) { epochInterval = genEpochInterval(r) },
	)
	params := types.NewParams(epochInterval)
	epochingGenesis := types.NewGenesis(params)

	bz, err := json.MarshalIndent(&epochingGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated epoching parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(epochingGenesis)
}
