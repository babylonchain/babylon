package keeper

import (
	"context"
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		blkHeight, fpBTCPK, err := parseBlkHeightAndPubKeyFromStoreKey(iter.Key())
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

// commitedRandoms iterates over all commited randoms on the store, parses the finality provider public key
// and the height from the iterator key and the commited random from the iterator value.
// This function has high resource consumption and should be only used on export genesis.
func (k Keeper) commitedRandoms(ctx context.Context) ([]*types.PublicRandomness, error) {
	store := k.pubRandStore(ctx)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	commtRandoms := make([]*types.PublicRandomness, 0)
	for ; iter.Valid(); iter.Next() {
		// key contains the fp and the block height
		fpBTCPK, blkHeight, err := parsePubKeyAndBlkHeightFromStoreKey(iter.Key())
		if err != nil {
			return nil, err
		}
		pubRand, err := bbn.NewSchnorrPubRand(iter.Value())
		if err != nil {
			return nil, err
		}

		commtRandoms = append(commtRandoms, &types.PublicRandomness{
			BlockHeight: blkHeight,
			FpBtcPk:     fpBTCPK,
			PubRand:     pubRand,
		})
	}

	return commtRandoms, nil
}

// parseBlkHeightAndPubKeyFromStoreKey expects to receive a key with
// BigEndianUint64(blkHeight) || BIP340PubKey(fpBTCPK)
func parseBlkHeightAndPubKeyFromStoreKey(key []byte) (blkHeight uint64, fpBTCPK *bbn.BIP340PubKey, err error) {
	sizeBigEndian := 8
	if len(key) < sizeBigEndian+1 {
		return 0, nil, fmt.Errorf("key not long enough to parse block height and BIP340PubKey: %s", key)
	}

	fpBTCPK, err = bbn.NewBIP340PubKey(key[sizeBigEndian:])
	if err != nil {
		return 0, nil, fmt.Errorf("failed to parse pub key from key %w: %w", bbn.ErrUnmarshal, err)
	}

	blkHeight = sdk.BigEndianToUint64(key[:sizeBigEndian])
	return blkHeight, fpBTCPK, nil
}

// parsePubKeyAndBlkHeightFromStoreKey expects to receive a key with
// BIP340PubKey(fpBTCPK) || BigEndianUint64(blkHeight)
func parsePubKeyAndBlkHeightFromStoreKey(key []byte) (fpBTCPK *bbn.BIP340PubKey, blkHeight uint64, err error) {
	sizeBigEndian := 8
	keyLen := len(key)
	if keyLen < sizeBigEndian+1 {
		return nil, 0, fmt.Errorf("key not long enough to parse BIP340PubKey and block height: %s", key)
	}

	startKeyHeight := keyLen - sizeBigEndian
	fpBTCPK, err = bbn.NewBIP340PubKey(key[:startKeyHeight])
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse pub key from key %w: %w", bbn.ErrUnmarshal, err)
	}

	blkHeight = sdk.BigEndianToUint64(key[startKeyHeight:])
	return fpBTCPK, blkHeight, nil
}
