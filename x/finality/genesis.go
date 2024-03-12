package finality

import (
	"context"

	bbn "github.com/babylonchain/babylon/types"

	"github.com/babylonchain/babylon/x/finality/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx context.Context, k keeper.Keeper, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(err)
	}

	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	for _, idxBlock := range gs.IndexedBlocks {
		k.SetBlock(ctx, idxBlock)
	}

	for _, evidence := range gs.Evidences {
		k.SetEvidence(ctx, evidence)
	}

	for _, voteSig := range gs.VoteSigs {
		k.SetSig(ctx, voteSig.BlockHeight, voteSig.FpBtcPk, voteSig.FinalitySig)
	}

	for _, commitRand := range gs.CommitedRandoms {
		// TODO: optimize insert?
		k.SetPubRandList(ctx, commitRand.FpBtcPk, commitRand.BlockHeight, []bbn.SchnorrPubRand{*commitRand.PubRand})
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	// TODO: get IndexedBlocks, Evidences, VoteSigs, CommitedRandoms
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
