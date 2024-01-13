package btclightclient

import (
	"context"

	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetBaseBTCHeader(ctx, genState.BaseBtcHeader)
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	baseBTCHeader := k.GetBaseBTCHeader(ctx)
	if baseBTCHeader == nil {
		panic("A base BTC Header has not been set")
	}

	genesis.BaseBtcHeader = *baseBTCHeader
	genesis.Params = k.GetParams(ctx)

	return genesis
}
