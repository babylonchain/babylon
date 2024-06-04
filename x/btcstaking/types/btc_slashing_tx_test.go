package types_test

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

// FuzzSlashingTx_VerifySigAndASig ensures properly generated adaptor signatures and
// the corresponding Schnorr signatures can be verified as valid
func FuzzSlashingTx_VerifySigAndASig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		var (
			stakingValue      = int64(2 * 10e8)
			stakingTimeBlocks = uint16(5)
			net               = &chaincfg.SimNetParams
		)

		// slashing address and key pairs
		slashingAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)

		// restaked to a random number of finality providers
		numRestakedFPs := int(datagen.RandomInt(r, 10) + 1)
		fpSKs, fpPKs, err := datagen.GenRandomBTCKeyPairs(r, numRestakedFPs)
		require.NoError(t, err)

		// use a random fp SK/PK
		fpIdx := int(datagen.RandomInt(r, numRestakedFPs))
		fpSK, fpPK := fpSKs[fpIdx], fpPKs[fpIdx]
		decKey, err := asig.NewDecyptionKeyFromBTCSK(fpSK)
		require.NoError(t, err)
		encKey, err := asig.NewEncryptionKeyFromBTCPK(fpPK)
		require.NoError(t, err)

		delSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		// (3, 5) covenant committee
		covenantSKs, covenantPKs, err := datagen.GenRandomBTCKeyPairs(r, 5)
		require.NoError(t, err)
		covenantQuorum := uint32(3)
		slashingChangeLockTime := uint16(101)

		// use a random covenant SK/PK
		covIdx := int(datagen.RandomInt(r, 5))
		covSK, covPK := covenantSKs[covIdx], covenantPKs[covIdx]

		// generate staking/slashing tx
		testStakingInfo := datagen.GenBTCStakingSlashingInfo(
			r,
			t,
			net,
			delSK,
			fpPKs,
			covenantPKs,
			covenantQuorum,
			stakingTimeBlocks,
			stakingValue,
			slashingAddress.EncodeAddress(),
			slashingRate,
			slashingChangeLockTime,
		)

		stakingTx := testStakingInfo.StakingTx
		slashingTx := testStakingInfo.SlashingTx

		slashingSpendInfo, err := testStakingInfo.StakingInfo.SlashingPathSpendInfo()
		require.NoError(t, err)
		slashingPkScriptPath := slashingSpendInfo.GetPkScriptPath()

		// ensure covenant adaptor signature can be correctly generated and verified
		covASig, err := slashingTx.EncSign(stakingTx, 0, slashingPkScriptPath, covSK, encKey)
		require.NoError(t, err)
		err = slashingTx.EncVerifyAdaptorSignature(
			testStakingInfo.StakingInfo.StakingOutput,
			slashingPkScriptPath,
			covPK,
			encKey,
			covASig,
		)
		require.NoError(t, err)

		// decrypt covenant adaptor signature and ensure the resulting Schnorr signature
		// can be verified
		covSig := covASig.Decrypt(decKey)
		err = slashingTx.VerifySignature(
			testStakingInfo.StakingInfo.StakingOutput,
			slashingPkScriptPath,
			covPK,
			bbn.NewBIP340SignatureFromBTCSig(covSig),
		)
		require.NoError(t, err)
	})
}

func FuzzSlashingTxWithWitness(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		var (
			stakingValue      = int64(2 * 10e8)
			stakingTimeBlocks = uint16(5)
			net               = &chaincfg.SimNetParams
		)

		// slashing address and key pairs
		slashingAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)

		// restaked to a random number of finality providers
		numRestakedFPs := int(datagen.RandomInt(r, 10) + 1)
		fpSKs, fpPKs, err := datagen.GenRandomBTCKeyPairs(r, numRestakedFPs)
		require.NoError(t, err)
		fpBTCPKs := bbn.NewBIP340PKsFromBTCPKs(fpPKs)

		// a random finality provider gets slashed
		fpIdx := int(datagen.RandomInt(r, numRestakedFPs))
		fpSK, fpPK := fpSKs[fpIdx], fpPKs[fpIdx]
		encKey, err := asig.NewEncryptionKeyFromBTCPK(fpPK)
		require.NoError(t, err)
		decKey, err := asig.NewDecyptionKeyFromBTCSK(fpSK)
		require.NoError(t, err)

		delSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		// (3, 5) covenant committee
		covenantSKs, covenantPKs, err := datagen.GenRandomBTCKeyPairs(r, 5)
		require.NoError(t, err)
		covenantQuorum := uint32(3)
		bsParams := types.Params{
			CovenantPks:    bbn.NewBIP340PKsFromBTCPKs(covenantPKs),
			CovenantQuorum: covenantQuorum,
		}
		slashingChangeLockTime := uint16(101)

		// generate staking/slashing tx
		testStakingInfo := datagen.GenBTCStakingSlashingInfo(
			r,
			t,
			net,
			delSK,
			fpPKs,
			covenantPKs,
			covenantQuorum,
			stakingTimeBlocks,
			stakingValue,
			slashingAddress.EncodeAddress(),
			slashingRate,
			slashingChangeLockTime,
		)

		slashingTx := testStakingInfo.SlashingTx
		stakingMsgTx := testStakingInfo.StakingTx

		slashingSpendInfo, err := testStakingInfo.StakingInfo.SlashingPathSpendInfo()
		require.NoError(t, err)
		slashingPkScriptPath := slashingSpendInfo.GetPkScriptPath()

		// delegator signs slashing tx
		delSig, err := slashingTx.Sign(stakingMsgTx, 0, slashingPkScriptPath, delSK)
		require.NoError(t, err)

		covenantSigners := covenantSKs[:covenantQuorum]
		// get covenant Schnorr signatures
		covenantSigs, err := datagen.GenCovenantAdaptorSigs(
			covenantSigners,
			fpPKs,
			stakingMsgTx,
			slashingPkScriptPath,
			slashingTx,
		)
		require.NoError(t, err)
		covSigsForFP, err := types.GetOrderedCovenantSignatures(fpIdx, covenantSigs, &bsParams)
		require.NoError(t, err)

		// ensure all covenant signatures encrypted by the slashed
		// finality provider's PK are verified
		orderedCovenantPKs := bbn.SortBIP340PKs(bsParams.CovenantPks)
		for i := range covSigsForFP {
			if covSigsForFP[i] == nil {
				continue
			}

			err := slashingTx.EncVerifyAdaptorSignature(
				testStakingInfo.StakingInfo.StakingOutput,
				slashingPkScriptPath,
				orderedCovenantPKs[i].MustToBTCPK(),
				encKey,
				covSigsForFP[i],
			)
			require.NoError(t, err, "verifying covenant adaptor sig at %d", i)

			covSchnorrSig := covSigsForFP[i].Decrypt(decKey)
			err = slashingTx.VerifySignature(
				testStakingInfo.StakingInfo.StakingOutput,
				slashingPkScriptPath,
				orderedCovenantPKs[i].MustToBTCPK(),
				bbn.NewBIP340SignatureFromBTCSig(covSchnorrSig),
			)
			require.NoError(t, err, "verifying covenant Schnorr sig at %d", i)
		}

		// create slashing tx with witness
		slashingMsgTxWithWitness, err := slashingTx.BuildSlashingTxWithWitness(fpSK, fpBTCPKs, stakingMsgTx, 0, delSig, covSigsForFP, slashingSpendInfo)
		require.NoError(t, err)

		// verify slashing tx with witness
		btctest.AssertSlashingTxExecution(t, testStakingInfo.StakingInfo.StakingOutput, slashingMsgTxWithWitness)
	})
}
