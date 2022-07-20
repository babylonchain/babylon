package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/babylonchain/babylon/x/epoching/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyEpochInterval),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", genEpochInterval(r))
			},
		),
	}
}
