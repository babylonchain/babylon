package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/boljen/go-bitmap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

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
// 6. (TODO: simulate blocks within these epochs, and generate inclusion proofs for valset and epoch metadata)
// 7. Execute VerifyEpochSealed with above epoch, checkpoint and proof, and assert the outcome to be true
//
// Tested property: proof is valid only when
// - BLS sig in proof is valid
// - TODO: BLS val set has a valid inclusion proof
// - TODO: epoch metadata has a valid inclusion proof
func FuzzProofEpochSealed(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// generate a random epoch
		epoch := datagen.GenRandomEpoch()

		// generate a random validator set with 100 validators
		numVals := 100
		valSet, blsSKs := datagen.GenerateValidatorSetWithBLSPrivKeys(numVals)

		// sample a validator subset, which may or may not reach a quorum
		bm, numSubSet := datagen.GenRandomBitmap()
		_, subsetPower, err := valSet.FindSubsetWithPowerSum(bm)
		require.NoError(t, err)

		// construct the rawCkpt
		// Note that the BlsMultiSig will be generated and assigned later
		lch := checkpointingtypes.LastCommitHash(epoch.SealerHeader.LastCommitHash)
		rawCkpt := &checkpointingtypes.RawCheckpoint{
			EpochNum:       epoch.EpochNumber,
			LastCommitHash: &lch,
			Bitmap:         bm,
			BlsMultiSig:    nil,
		}

		// let the subset generate a BLS multisig over sealer header's last_commit_hash
		multiSig, err := signBLSWithBitmap(blsSKs, bm, rawCkpt.SignedMsg())
		require.NoError(t, err)
		// assign multiSig to rawCkpt
		rawCkpt.BlsMultiSig = &multiSig

		// mock checkpointing keeper that produces the expected validator set
		checkpointingKeeper := zctypes.NewMockCheckpointingKeeper(ctrl)
		checkpointingKeeper.EXPECT().GetBLSPubKeySet(gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return(valSet, nil).AnyTimes()
		// create zcKeeper and ctx
		zcKeeper, ctx := testkeeper.ZoneConciergeKeeper(t, checkpointingKeeper, nil, nil, nil)

		// prove
		proof, err := zcKeeper.ProveEpochSealed(ctx, epoch.EpochNumber)
		require.NoError(t, err)
		// verify
		err = zckeeper.VerifyEpochSealed(epoch, rawCkpt, proof)

		if numSubSet <= numVals*1/3 { // BLS sig does not reach a quorum
			require.LessOrEqual(t, subsetPower, uint64(numVals*1/3))
			require.Error(t, err)
		} else { // BLS sig has a valid quorum
			require.Greater(t, subsetPower, valSet.GetTotalPower()*1/3)
			require.NoError(t, err)
		}
	})
}
