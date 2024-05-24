package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/stretchr/testify/require"
)

func TestExportGenesis(t *testing.T) {
	k, ctx := keepertest.FinalityKeeper(t, nil, nil)

	r := rand.New(rand.NewSource(10))
	btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)

	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	blkHeight, startHeight, numPubRand := uint64(1), uint64(0), uint64(5)

	randListInfo, _, err := datagen.GenRandomMsgCommitPubRandList(r, btcSK, startHeight, numPubRand)
	require.NoError(t, err)

	blockHash := datagen.GenRandomByteArray(r, 32)
	signer := datagen.GenRandomAccount().Address
	msgAddFinalitySig, err := datagen.NewMsgAddFinalitySig(signer, btcSK, startHeight, blkHeight, randListInfo, blockHash)
	require.NoError(t, err)

	allVotes := make([]*types.VoteSig, numPubRand)
	allBlocks := make([]*types.IndexedBlock, numPubRand)
	allEvidences := make([]*types.Evidence, numPubRand)
	allPublicRandomness := make([]*types.PublicRandomness, numPubRand)
	for i := 0; i < int(numPubRand); i++ {
		// Votes
		vt := &types.VoteSig{
			FpBtcPk:     fpBTCPK,
			BlockHeight: blkHeight,
			FinalitySig: msgAddFinalitySig.FinalitySig,
		}
		k.SetSig(ctx, vt.BlockHeight, vt.FpBtcPk, vt.FinalitySig)
		allVotes[i] = vt

		// Blocks
		blk := &types.IndexedBlock{
			Height:    blkHeight,
			AppHash:   blockHash,
			Finalized: i%2 == 0,
		}
		k.SetBlock(ctx, blk)
		allBlocks[i] = blk

		// Evidences
		evidence := &types.Evidence{
			FpBtcPk:              fpBTCPK,
			BlockHeight:          blkHeight,
			PubRand:              &randListInfo.PRList[i],
			ForkAppHash:          msgAddFinalitySig.BlockAppHash,
			ForkFinalitySig:      msgAddFinalitySig.FinalitySig,
			CanonicalAppHash:     blockHash,
			CanonicalFinalitySig: msgAddFinalitySig.FinalitySig,
		}
		k.SetEvidence(ctx, evidence)
		allEvidences[i] = evidence

		// public randomness
		pubRand := randListInfo.PRList[i]
		k.SetPubRand(ctx, fpBTCPK, blkHeight, pubRand)
		randomness := &types.PublicRandomness{
			BlockHeight: blkHeight,
			FpBtcPk:     fpBTCPK,
			PubRand:     &pubRand,
		}
		allPublicRandomness[i] = randomness

		// updates the block everytime to make sure something is different.
		blkHeight++
	}

	prc := &types.PubRandCommit{
		StartHeight: startHeight,
		NumPubRand:  numPubRand,
		Commitment:  randListInfo.Commitment,
	}
	k.SetPubRandCommit(ctx, fpBTCPK, prc)

	require.Equal(t, len(allVotes), int(numPubRand))
	require.Equal(t, len(allBlocks), int(numPubRand))
	require.Equal(t, len(allEvidences), int(numPubRand))
	require.Equal(t, len(allPublicRandomness), int(numPubRand))

	gs, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Equal(t, k.GetParams(ctx), gs.Params)

	require.Equal(t, allVotes, gs.VoteSigs)
	require.Equal(t, allBlocks, gs.IndexedBlocks)
	require.Equal(t, allEvidences, gs.Evidences)
	require.Equal(t, allPublicRandomness, gs.PublicRandomness)
	require.Equal(t, prc, gs.PubRandCommit[0].PubRandCommit)
}
