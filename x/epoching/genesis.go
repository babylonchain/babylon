package epoching

import (
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// set params for this module
	k.SetParams(ctx, genState.Params)
	// init epoch number
	k.SetEpochNumber(ctx, sdk.NewUint(0))
	// init msg queue length
	k.SetQueueLength(ctx, sdk.NewUint(0))
	// init slashed validator set size
	k.InitSlashedValidatorSetSize(ctx)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
