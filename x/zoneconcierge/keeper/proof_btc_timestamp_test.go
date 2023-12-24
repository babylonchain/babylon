package keeper_test

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/boljen/go-bitmap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/testutil/datagen"
	testhelper "github.com/babylonchain/babylon/testutil/helper"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
)

func FuzzProofCZHeaderInEpoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		h := testhelper.NewHelper(t)
		ek := h.App.EpochingKeeper
		zck := h.App.ZoneConciergeKeeper
		var err error

		// chain is at height 1 thus epoch 1

		// enter the 1st block of epoch 2
		epochInterval := ek.GetParams(h.Ctx).EpochInterval
		for j := 0; j < int(epochInterval)-1; j++ {
			h.Ctx, err = h.ApplyEmptyBlockWithVoteExtension(r)
			h.NoError(err)
		}

		// handle a random header from a random consumer chain
		chainID := datagen.GenRandomHexStr(r, 10)
		height := datagen.RandomInt(r, 100) + 1
		ibctmHeader := datagen.GenRandomIBCTMHeader(r, chainID, height)
		headerInfo := datagen.HeaderToHeaderInfo(ibctmHeader)
		zck.HandleHeaderWithValidCommit(h.Ctx, datagen.GenRandomByteArray(r, 32), headerInfo, false)

		// ensure the header is successfully inserted
		indexedHeader, err := zck.GetHeader(h.Ctx, chainID, height)
		h.NoError(err)

		// enter the 1st block of the next epoch
		for j := 0; j < int(epochInterval); j++ {
			h.Ctx, err = h.ApplyEmptyBlockWithVoteExtension(r)
			h.NoError(err)
		}

		epochWithHeader, err := ek.GetHistoricalEpoch(h.Ctx, indexedHeader.BabylonEpoch)
		h.NoError(err)

		// generate inclusion proof
		proof, err := zck.ProveCZHeaderInEpoch(h.Ctx, indexedHeader, epochWithHeader)
		h.NoError(err)

		// verify the inclusion proof
		err = zctypes.VerifyCZHeaderInEpoch(indexedHeader, epochWithHeader, proof)
		h.NoError(err)
	})
}

func signBLSWithBitmap(blsSKs []bls12381.PrivateKey, bm bitmap.Bitmap, msg []byte) (bls12381.Signature, error) {
	sigs := []bls12381.Signature{}
	for i := 0; i < len(blsSKs); i++ {
		if bitmap.Get(bm, i) {
			sig := bls12381.Sign(blsSKs[i], msg)
			sigs = append(sigs, sig)
		}
	}
	return bls12381.AggrSigList(sigs)
}

// FuzzProofEpochSealed fuzz tests the prover and verifier of ProofEpochSealed
// Process:
// 1. Generate a random epoch that has a legitimate-looking SealerHeader
// 2. Generate a random validator set with BLS PKs
// 3. Generate a BLS multisig with >1/3 random validators of the validator set
// 4. Generate a checkpoint based on the above validator subset and the above sealer header
// 5. Execute ProveEpochSealed where the mocked checkpointing keeper produces the above validator set
// 6. Execute VerifyEpochSealed with above epoch, checkpoint and proof, and assert the outcome to be true
//
// Tested property: proof is valid only when
// - BLS sig in proof is valid
func FuzzProofEpochSealed_BLSSig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// generate a random epoch
		epoch := datagen.GenRandomEpoch(r)

		// generate a random validator set with 100 validators
		numVals := 100
		valSet, blsSKs := datagen.GenerateValidatorSetWithBLSPrivKeys(numVals)

		// sample a validator subset, which may or may not reach a quorum
		bm, numSubSet := datagen.GenRandomBitmap(r)
		_, subsetPower, err := valSet.FindSubsetWithPowerSum(bm)
		require.NoError(t, err)

		// construct the rawCkpt
		// Note that the BlsMultiSig will be generated and assigned later
		blockHash := checkpointingtypes.BlockHash(epoch.SealerBlockHash)
		rawCkpt := &checkpointingtypes.RawCheckpoint{
			EpochNum:    epoch.EpochNumber,
			BlockHash:   &blockHash,
			Bitmap:      bm,
			BlsMultiSig: nil,
		}

		// let the subset generate a BLS multisig over sealer header's app_hash
		multiSig, err := signBLSWithBitmap(blsSKs, bm, rawCkpt.SignedMsg())
		require.NoError(t, err)
		// assign multiSig to rawCkpt
		rawCkpt.BlsMultiSig = &multiSig

		// mock checkpointing keeper that produces the expected validator set
		checkpointingKeeper := zctypes.NewMockCheckpointingKeeper(ctrl)
		checkpointingKeeper.EXPECT().GetBLSPubKeySet(gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return(valSet.ValSet, nil).AnyTimes()
		// mock epoching keeper
		epochingKeeper := zctypes.NewMockEpochingKeeper(ctrl)
		epochingKeeper.EXPECT().GetEpoch(gomock.Any()).Return(epoch).AnyTimes()
		epochingKeeper.EXPECT().GetHistoricalEpoch(gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return(epoch, nil).AnyTimes()
		// create zcKeeper and ctx
		zcKeeper, ctx := testkeeper.ZoneConciergeKeeper(t, nil, checkpointingKeeper, nil, epochingKeeper)

		// prove
		proof, err := zcKeeper.ProveEpochSealed(ctx, epoch.EpochNumber)
		require.NoError(t, err)
		// verify
		err = zctypes.VerifyEpochSealed(epoch, rawCkpt, proof)

		if subsetPower*3 <= valSet.GetTotalPower()*2 { // BLS sig does not reach a quorum
			require.LessOrEqual(t, numSubSet, numVals*2/3)
			require.Error(t, err)
			require.NotErrorIs(t, err, zctypes.ErrInvalidMerkleProof)
		} else { // BLS sig has a valid quorum
			require.Greater(t, numSubSet, numVals*2/3)
			require.Error(t, err)
			require.ErrorIs(t, err, zctypes.ErrInvalidMerkleProof)
		}
	})
}

func FuzzProofEpochSealed_Epoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		h := testhelper.NewHelper(t)
		ek := h.App.EpochingKeeper
		zck := h.App.ZoneConciergeKeeper
		var err error

		// chain is at height 1

		// enter the 1st block of a random epoch
		epochInterval := ek.GetParams(h.Ctx).EpochInterval
		newEpochs := datagen.RandomInt(r, 10) + 2
		for i := 0; i < int(newEpochs); i++ {
			for j := 0; j < int(epochInterval); j++ {
				h.Ctx, err = h.ApplyEmptyBlockWithVoteExtension(r)
				h.NoError(err)
			}
		}

		// prove the inclusion of last epoch
		lastEpochNumber := ek.GetEpoch(h.Ctx).EpochNumber - 1
		h.NoError(err)
		lastEpoch, err := ek.GetHistoricalEpoch(h.Ctx, lastEpochNumber)
		h.NoError(err)
		proof, err := zck.ProveEpochInfo(lastEpoch)
		h.NoError(err)

		// verify inclusion proof
		err = zctypes.VerifyEpochInfo(lastEpoch, proof)
		h.NoError(err)
	})
}

func FuzzProofEpochSealed_ValSet(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// generate the validator set with 10 validators as genesis
		genesisValSet, privSigner, err := datagen.GenesisValidatorSetWithPrivSigner(10)
		require.NoError(t, err)
		h := testhelper.NewHelperWithValSet(t, genesisValSet, privSigner)
		ek := h.App.EpochingKeeper
		ck := h.App.CheckpointingKeeper
		zck := h.App.ZoneConciergeKeeper

		// chain is at height 1
		// enter the 1st block of a random epoch
		epochInterval := ek.GetParams(h.Ctx).EpochInterval
		newEpochs := datagen.RandomInt(r, 10) + 2
		for i := 0; i < int(newEpochs); i++ {
			for j := 0; j < int(epochInterval); j++ {
				_, err = h.ApplyEmptyBlockWithVoteExtension(r)
				h.NoError(err)
			}
		}

		// seal the last epoch at the 2nd block of the current epoch
		h.Ctx, err = h.ApplyEmptyBlockWithVoteExtension(r)
		h.NoError(err)

		// prove the inclusion of last epoch
		lastEpochNumber := ek.GetEpoch(h.Ctx).EpochNumber - 1
		h.NoError(err)
		lastEpoch, err := ek.GetHistoricalEpoch(h.Ctx, lastEpochNumber)
		h.NoError(err)
		lastEpochValSet := ck.GetValidatorBlsKeySet(h.Ctx, lastEpochNumber)
		proof, err := zck.ProveValSet(lastEpoch)
		h.NoError(err)

		// verify inclusion proof
		err = zctypes.VerifyValSet(lastEpoch, lastEpochValSet, proof)
		h.NoError(err)
	})
}

func FuzzProofEpochSubmitted(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// generate random epoch, random rawBtcCkpt and random rawCkpt
		epoch := datagen.GenRandomEpoch(r)
		rawBtcCkpt := datagen.GetRandomRawBtcCheckpoint(r)
		rawBtcCkpt.Epoch = epoch.EpochNumber
		rawCkpt, err := checkpointingtypes.FromBTCCkptToRawCkpt(rawBtcCkpt)
		require.NoError(t, err)

		// encode ckpt to BTC txs in BTC blocks
		testRawCkptData := datagen.EncodeRawCkptToTestData(rawBtcCkpt)
		idxs := []uint64{datagen.RandomInt(r, 5) + 1, datagen.RandomInt(r, 5) + 1}
		offsets := []uint64{datagen.RandomInt(r, 5) + 1, datagen.RandomInt(r, 5) + 1}
		btcBlocks := []*datagen.BlockCreationResult{
			datagen.CreateBlock(r, 1, uint32(idxs[0]+offsets[0]), uint32(idxs[0]), testRawCkptData.FirstPart),
			datagen.CreateBlock(r, 2, uint32(idxs[1]+offsets[1]), uint32(idxs[1]), testRawCkptData.SecondPart),
		}
		// create MsgInsertBtcSpvProof for the rawCkpt
		msgInsertBtcSpvProof := datagen.GenerateMessageWithRandomSubmitter([]*datagen.BlockCreationResult{btcBlocks[0], btcBlocks[1]})

		// get headers for verification
		btcHeaders := []*wire.BlockHeader{
			btcBlocks[0].HeaderBytes.ToBlockHeader(),
			btcBlocks[1].HeaderBytes.ToBlockHeader(),
		}

		// get 2 tx info for the ckpt parts
		txsInfo := []*btcctypes.TransactionInfo{
			{
				Key:         &btcctypes.TransactionKey{Index: uint32(idxs[0]), Hash: btcBlocks[0].HeaderBytes.Hash()},
				Transaction: msgInsertBtcSpvProof.Proofs[0].BtcTransaction,
				Proof:       msgInsertBtcSpvProof.Proofs[0].MerkleNodes,
			},
			{
				Key:         &btcctypes.TransactionKey{Index: uint32(idxs[1]), Hash: btcBlocks[1].HeaderBytes.Hash()},
				Transaction: msgInsertBtcSpvProof.Proofs[1].BtcTransaction,
				Proof:       msgInsertBtcSpvProof.Proofs[1].MerkleNodes,
			},
		}

		// net param, babylonTag
		powLimit := chaincfg.SimNetParams.PowLimit
		babylonTag := btcctypes.DefaultCheckpointTag
		tagAsBytes, _ := hex.DecodeString(babylonTag)

		// verify
		err = zctypes.VerifyEpochSubmitted(rawCkpt, txsInfo, btcHeaders, powLimit, tagAsBytes)
		require.NoError(t, err)
	})
}
