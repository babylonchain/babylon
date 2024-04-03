package checkpointing

import (
	"context"

	"github.com/babylonchain/babylon/x/checkpointing/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetGenBlsKeys(ctx, genState.GenesisKeys)
	// set epoch 0 to be finalised at genesis
	k.SetLastFinalizedEpoch(ctx, 0)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	return genesis
}
