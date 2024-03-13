package types_test

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"

	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

func FuzzBTCDelegation(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		btcDel := &types.BTCDelegation{}
		// randomise voting power
		btcDel.TotalSat = datagen.RandomInt(r, 100000)
		btcDel.BtcUndelegation = &types.BTCUndelegation{}

		// randomise covenant sig
		hasCovenantSig := datagen.RandomInt(r, 2) == 0
		if hasCovenantSig {
			encKey, _, err := asig.GenKeyPair()
			require.NoError(t, err)
			covenantSK, _ := btcec.PrivKeyFromBytes(
				[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			)
			covenantSig, err := asig.EncSign(covenantSK, encKey, datagen.GenRandomByteArray(r, 32))
			require.NoError(t, err)
			covPk, err := datagen.GenRandomBIP340PubKey(r)
			require.NoError(t, err)
			covSigInfo := &types.CovenantAdaptorSignatures{
				CovPk:       covPk,
				AdaptorSigs: [][]byte{covenantSig.MustMarshal()},
			}
			btcDel.CovenantSigs = []*types.CovenantAdaptorSignatures{covSigInfo}
			btcDel.BtcUndelegation.CovenantSlashingSigs = btcDel.CovenantSigs                                // doesn't matter
			btcDel.BtcUndelegation.CovenantUnbondingSigList = []*types.SignatureInfo{&types.SignatureInfo{}} // doesn't matter
		}

		// randomise start height and end height
		btcDel.StartHeight = datagen.RandomInt(r, 100)
		btcDel.EndHeight = btcDel.StartHeight + datagen.RandomInt(r, 100)

		// randomise BTC tip and w
		btcHeight := btcDel.StartHeight + datagen.RandomInt(r, 50)
		w := datagen.RandomInt(r, 50)

		// test expected voting power
		hasVotingPower := hasCovenantSig && btcDel.StartHeight <= btcHeight && btcHeight+w <= btcDel.EndHeight
		actualVotingPower := btcDel.VotingPower(btcHeight, w, 1)
		if hasVotingPower {
			require.Equal(t, btcDel.TotalSat, actualVotingPower)
		} else {
			require.Equal(t, uint64(0), actualVotingPower)
		}
	})
}

func FuzzBTCDelegation_SlashingTx(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		net := &chaincfg.SimNetParams

		delSK, delPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		delBTCPK := bbn.NewBIP340PubKeyFromBTCPK(delPK)

		// restaked to a random number of finality providers
		numRestakedFPs := int(datagen.RandomInt(r, 10) + 1)
		fpSKs, fpPKs, err := datagen.GenRandomBTCKeyPairs(r, numRestakedFPs)
		require.NoError(t, err)

		// (3, 5) covenant committee
		covenantSKs, covenantPKs, err := datagen.GenRandomBTCKeyPairs(r, 5)
		require.NoError(t, err)
		covenantQuorum := uint32(3)

		stakingTimeBlocks := uint16(5)
		stakingValue := int64(2 * 10e8)
		slashingAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)

		slashingChangeLockTime := uint16(101)

		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)
		testInfo := datagen.GenBTCStakingSlashingInfo(
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
		require.NoError(t, err)

		stakingTxBytes, err := bbn.SerializeBTCTx(testInfo.StakingTx)
		require.NoError(t, err)

		// spend info of the slashing tx
		slashingSpendInfo, err := testInfo.StakingInfo.SlashingPathSpendInfo()
		require.NoError(t, err)
		// delegator signs the slashing tx
		delSig, err := testInfo.SlashingTx.Sign(testInfo.StakingTx, 0, slashingSpendInfo.GetPkScriptPath(), delSK)
		require.NoError(t, err)
		// covenant signs (using adaptor signature) the slashing tx
		covenantSigs, err := datagen.GenCovenantAdaptorSigs(
			covenantSKs,
			fpPKs,
			testInfo.StakingTx,
			slashingSpendInfo.GetPkScriptPath(),
			testInfo.SlashingTx,
		)
		require.NoError(t, err)
		covenantSigs = covenantSigs[2:] // discard 2 out of 5 signatures

		// construct the BTC delegation with everything
		btcDel := &types.BTCDelegation{
			BabylonPk:        nil, // not relevant here
			BtcPk:            delBTCPK,
			Pop:              nil, // not relevant here
			FpBtcPkList:      bbn.NewBIP340PKsFromBTCPKs(fpPKs),
			StartHeight:      1000, // not relevant here
			EndHeight:        uint64(1000 + stakingTimeBlocks),
			TotalSat:         uint64(stakingValue),
			StakingTx:        stakingTxBytes,
			StakingOutputIdx: 0,
			SlashingTx:       testInfo.SlashingTx,
			DelegatorSig:     delSig,
			CovenantSigs:     covenantSigs,
		}

		bsParams := &types.Params{
			CovenantPks:    bbn.NewBIP340PKsFromBTCPKs(covenantPKs),
			CovenantQuorum: covenantQuorum,
		}
		btcNet := &chaincfg.SimNetParams

		// build slashing tx with witness for spending the staking tx
		// a random finality provider gets slashed
		slashedFPIdx := int(datagen.RandomInt(r, numRestakedFPs))
		fpSK := fpSKs[slashedFPIdx]
		slashingTxWithWitness, err := btcDel.BuildSlashingTxWithWitness(bsParams, btcNet, fpSK)
		require.NoError(t, err)

		// assert execution
		btctest.AssertSlashingTxExecution(t, testInfo.StakingInfo.StakingOutput, slashingTxWithWitness)
	})
}
