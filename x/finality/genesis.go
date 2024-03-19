package finality

import (
	"context"

	"github.com/babylonchain/babylon/x/finality/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx context.Context, k keeper.Keeper, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(err)
	}

	if err := k.InitGenesis(ctx, gs); err != nil {
		panic(err)
	}
	// SetVotes
	// SetPubRand
	// SetEvidence
	// SetNextHeightToFinalize
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	gs, err := k.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return gs
}
