package simulation

// DONTCOVER

import (
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals on the simulation
// TODO: add support of changing EpochInterval on-the-fly
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		// simulation.NewSimParamChange(types.ModuleName, string(types.KeyEpochInterval),
		// 	func(r *rand.Rand) string {
		// 		return fmt.Sprintf("%d", genEpochInterval(r))
		// 	},
		// ),
	}
}
