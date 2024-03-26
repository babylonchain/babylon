package keeper

import (
	"context"

	btcstk "github.com/babylonchain/babylon/btcstaking"
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

	return k.SetParams(ctx, gs.Params)
}

// ExportGenesis returns the keeper state into a exported genesis state.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
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

	return &types.GenesisState{
		Params:        k.GetParams(ctx),
		IndexedBlocks: blocks,
		Evidences:     evidences,
		VoteSigs:      voteSigs,
	}, nil
}

// blocks loads all blocks stored.
// This function has high resource consumption and should be only used on export genesis.
func (k Keeper) blocks(ctx context.Context) ([]*types.IndexedBlock, error) {
	blocks := make([]*types.IndexedBlock, 0)

	iter := k.blockStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var blk types.IndexedBlock
		if err := k.cdc.Unmarshal(iter.Value(), &blk); err != nil {
			return nil, err
		}
		blocks = append(blocks, &blk)
	}

	return blocks, nil
}

// evidences loads all evidences stored.
// This function has high resource consumption and should be only used on export genesis.
func (k Keeper) evidences(ctx context.Context) (evidences []*types.Evidence, err error) {
	evidences = make([]*types.Evidence, 0)

	iter := k.evidenceStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var evd types.Evidence
		if err := k.cdc.Unmarshal(iter.Value(), &evd); err != nil {
			return nil, err
		}
		evidences = append(evidences, &evd)
	}

	return evidences, nil
}

// voteSigs iterates over all votes on the store, parses the height and the finality provider
// public key from the iterator key and the finality signature from the iterator value.
// This function has high resource consumption and should be only used on export genesis.
func (k Keeper) voteSigs(ctx context.Context) ([]*types.VoteSig, error) {
	store := k.voteStore(ctx)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	voteSigs := make([]*types.VoteSig, 0)
	for ; iter.Valid(); iter.Next() {
		// key contains the height and the fp
		blkHeight, fpBTCPK, err := btcstk.ParseBlkHeightAndPubKeyFromStoreKey(iter.Key())
		if err != nil {
			return nil, err
		}
		finalitySig, err := bbn.NewSchnorrEOTSSig(iter.Value())
		if err != nil {
			return nil, err
		}

		voteSigs = append(voteSigs, &types.VoteSig{
			BlockHeight: blkHeight,
			FpBtcPk:     fpBTCPK,
			FinalitySig: finalitySig,
		})
	}

	return voteSigs, nil
}
