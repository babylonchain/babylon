package epoching

import (
	"context"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	// set params for this module
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// init epoch number
	k.InitEpoch(ctx)
	// init msg queue
	k.InitMsgQueue(ctx)
	// init validator set
	k.InitValidatorSet(ctx)
	// init slashed voting power
	k.InitSlashedVotingPower(ctx)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
