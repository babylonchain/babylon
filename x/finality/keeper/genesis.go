package keeper

import (
	"context"
	"fmt"

	btcstk "github.com/babylonchain/babylon/btcstaking"
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

	for _, pubRand := range gs.PublicRandomness {
		k.SetPubRandList(ctx, pubRand.FpBtcPk, pubRand.BlockHeight, []bbn.SchnorrPubRand{*pubRand.PubRand})
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

	pubRandomness, err := k.publicRandomness(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GenesisState{
		Params:           k.GetParams(ctx),
		IndexedBlocks:    blocks,
		Evidences:        evidences,
		VoteSigs:         voteSigs,
		PublicRandomness: pubRandomness,
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

// publicRandomness iterates over all commited randoms on the store, parses the finality provider public key
// and the height from the iterator key and the commited random from the iterator value.
// This function has high resource consumption and should be only used on export genesis.
func (k Keeper) publicRandomness(ctx context.Context) ([]*types.PublicRandomness, error) {
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
