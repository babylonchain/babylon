package keeper

import (
	"context"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
)

// InitGenesis initializes the keeper state from a provided initial genesis state.
func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
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
		k.SetPubRandList(ctx, commitRand.FpBtcPk, commitRand.BlockHeight, []bbn.SchnorrPubRand{*commitRand.PubRand})
	}

	return k.SetParams(ctx, gs.Params)
}
