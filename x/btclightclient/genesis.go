package btclightclient

import (
	"context"

	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(err)
	}

	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	k.InsertHeaderInfos(ctx, gs.BtcHeaders)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:     k.GetParams(ctx),
		BtcHeaders: k.GetAllHeaderInfos(ctx),
	}
}
