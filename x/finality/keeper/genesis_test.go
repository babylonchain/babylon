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

func TestExportGenesisCheckEvidences(t *testing.T) {
	k, ctx := keepertest.FinalityKeeper(t, nil, nil)

	r := rand.New(rand.NewSource(10))
	btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)

	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	blkHeight, startHeight, numPubRand := uint64(1), uint64(0), uint64(5)

	srList, msgCommitPubRandList, err := datagen.GenRandomMsgCommitPubRandList(r, btcSK, startHeight, numPubRand)
	require.NoError(t, err)

	sr := srList[startHeight+blkHeight]
	blockHash := datagen.GenRandomByteArray(r, 32)
	signer := datagen.GenRandomAccount().Address
	msgAddFinalitySig, err := types.NewMsgAddFinalitySig(signer, btcSK, sr, blkHeight, blockHash)
	require.NoError(t, err)

	allEvidences := make([]*types.Evidence, numPubRand)
	for i := 0; i < int(numPubRand); i++ {
		evidence := &types.Evidence{
			FpBtcPk:              fpBTCPK,
			BlockHeight:          blkHeight,
			PubRand:              &msgCommitPubRandList.PubRandList[i],
			ForkAppHash:          msgAddFinalitySig.BlockAppHash,
			ForkFinalitySig:      msgAddFinalitySig.FinalitySig,
			CanonicalAppHash:     blockHash,
			CanonicalFinalitySig: msgAddFinalitySig.FinalitySig,
		}
		k.SetEvidence(ctx, evidence)
		allEvidences[i] = evidence
		blkHeight++
	}

	gs, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Equal(t, allEvidences, gs.Evidences)
}

func TestExportGenesisCheckVoteSigs(t *testing.T) {
	k, ctx := keepertest.FinalityKeeper(t, nil, nil)
	r := rand.New(rand.NewSource(10))

	btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)

	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	blkHeight, startHeight, numPubRand := uint64(1), uint64(0), uint64(5)

	srList, _, err := datagen.GenRandomMsgCommitPubRandList(r, btcSK, startHeight, numPubRand)
	require.NoError(t, err)

	sr := srList[startHeight+blkHeight]
	blockHash := datagen.GenRandomByteArray(r, 32)
	signer := datagen.GenRandomAccount().Address
	msgAddFinalitySig, err := types.NewMsgAddFinalitySig(signer, btcSK, sr, blkHeight, blockHash)
	require.NoError(t, err)

	allVotes := make([]*types.VoteSig, numPubRand)
	for i := 0; i < int(numPubRand); i++ {
		vt := &types.VoteSig{
			FpBtcPk:     fpBTCPK,
			BlockHeight: blkHeight,
			FinalitySig: msgAddFinalitySig.FinalitySig,
		}
		k.SetSig(ctx, vt.BlockHeight, vt.FpBtcPk, vt.FinalitySig)
		allVotes[i] = vt
		blkHeight++
	}

	gs, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Equal(t, allVotes, gs.VoteSigs)
}
