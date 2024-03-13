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

// ExportGenesis returns the keeper state into a exported genesis state.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	// TODO: get VoteSigs, CommitedRandoms
	blocks, err := k.blocks(ctx)
	if err != nil {
		return nil, err
	}

	evidences, err := k.evidences(ctx)
	if err != nil {
		return nil, err
	}

	voteSigs, err := k.voteSigs(ctx)
	if err != nil {
		return nil, err
	}

	commitedRandoms, err := k.commitedRandoms(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GenesisState{
		Params:          k.GetParams(ctx),
		IndexedBlocks:   blocks,
		Evidences:       evidences,
		VoteSigs:        voteSigs,
		CommitedRandoms: commitedRandoms,
	}, nil
}
