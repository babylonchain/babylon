package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/eots"
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
	msr, mpr, err := eots.NewMasterRandPair(r)
	require.NoError(t, err)

	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	blkHeight, startHeight, numPubRand := uint64(1), uint64(0), uint64(5)

	sr, _, err := msr.DeriveRandPair(uint32(startHeight + blkHeight))
	require.NoError(t, err)
	blockHash := datagen.GenRandomByteArray(r, 32)
	signer := datagen.GenRandomAccount().Address
	msgAddFinalitySig, err := types.NewMsgAddFinalitySig(signer, btcSK, sr, blkHeight, blockHash)
	require.NoError(t, err)

	allVotes := make([]*types.VoteSig, numPubRand)
	allBlocks := make([]*types.IndexedBlock, numPubRand)
	allEvidences := make([]*types.Evidence, numPubRand)
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
			ForkAppHash:          msgAddFinalitySig.BlockAppHash,
			ForkFinalitySig:      msgAddFinalitySig.FinalitySig,
			CanonicalAppHash:     blockHash,
			CanonicalFinalitySig: msgAddFinalitySig.FinalitySig,
			MasterPubRand:        mpr.MarshalBase58(),
		}
		k.SetEvidence(ctx, evidence)
		allEvidences[i] = evidence

		// updates the block everytime to make sure something is different.
		blkHeight++
	}
	require.Equal(t, len(allVotes), int(numPubRand))
	require.Equal(t, len(allBlocks), int(numPubRand))
	require.Equal(t, len(allEvidences), int(numPubRand))

	gs, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Equal(t, k.GetParams(ctx), gs.Params)

	require.Equal(t, allVotes, gs.VoteSigs)
	require.Equal(t, allBlocks, gs.IndexedBlocks)
	require.Equal(t, allEvidences, gs.Evidences)
}
