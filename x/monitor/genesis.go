package monitor

import (
	"context"
	"github.com/babylonchain/babylon/x/monitor/keeper"
	"github.com/babylonchain/babylon/x/monitor/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	return genesis
}
