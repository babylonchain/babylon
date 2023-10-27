package btclightclient

import (
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetBaseBTCHeader(ctx, genState.BaseBtcHeader)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	baseBTCHeader := k.GetBaseBTCHeader(ctx)
	if baseBTCHeader == nil {
		panic("A base BTC Header has not been set")
	}
	genesis.BaseBtcHeader = *baseBTCHeader

	return genesis
}
