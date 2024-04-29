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

		delSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		// restaked to a random number of finality providers
		numRestakedFPs := int(datagen.RandomInt(r, 10) + 1)
		fpSKs, fpPKs, err := datagen.GenRandomBTCKeyPairs(r, numRestakedFPs)
		require.NoError(t, err)
		fpBTCPKs := bbn.NewBIP340PKsFromBTCPKs(fpPKs)

		// a random finality provider gets slashed
		slashedFPIdx := int(datagen.RandomInt(r, numRestakedFPs))
		fpSK := fpSKs[slashedFPIdx]

		// (3, 5) covenant committee
		covenantSKs, covenantPKs, err := datagen.GenRandomBTCKeyPairs(r, 5)
		require.NoError(t, err)
		covenantQuorum := uint32(3)
		bsParams := &types.Params{
			CovenantPks:    bbn.NewBIP340PKsFromBTCPKs(covenantPKs),
			CovenantQuorum: covenantQuorum,
		}

		stakingTimeBlocks := uint16(5)
		stakingValue := int64(2 * 10e8)
		slashingAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)

		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)
		unbondingTime := uint16(100) + 1
		slashingChangeLockTime := unbondingTime

		// only the quorum of signers provided the signatures
		covenantSigners := covenantSKs[:covenantQuorum]

		// construct the BTC delegation with everything
		btcDel, err := datagen.GenRandomBTCDelegation(
			r,
			t,
			&chaincfg.SimNetParams,
			fpBTCPKs,
			delSK,
			covenantSigners,
			covenantPKs,
			covenantQuorum,
			slashingAddress.EncodeAddress(),
			1000,
			uint64(1000+stakingTimeBlocks),
			uint64(stakingValue),
			slashingRate,
			slashingChangeLockTime,
		)
		require.NoError(t, err)

		stakingInfo, err := btcDel.GetStakingInfo(bsParams, net)
		require.NoError(t, err)

		// TESTING
		orderedCovenantPKs := bbn.SortBIP340PKs(bsParams.CovenantPks)
		covSigsForFP, err := types.GetOrderedCovenantSignatures(slashedFPIdx, btcDel.CovenantSigs, bsParams)
		require.NoError(t, err)
		fpPK := fpSK.PubKey()
		encKey, err := asig.NewEncryptionKeyFromBTCPK(fpPK)
		require.NoError(t, err)
		slashingSpendInfo, err := stakingInfo.SlashingPathSpendInfo()
		require.NoError(t, err)
		for i := range covSigsForFP {
			if covSigsForFP[i] == nil {
				continue
			}
			err := btcDel.SlashingTx.EncVerifyAdaptorSignature(
				stakingInfo.StakingOutput,
				slashingSpendInfo.GetPkScriptPath(),
				orderedCovenantPKs[i].MustToBTCPK(),
				encKey,
				covSigsForFP[i],
			)
			require.NoError(t, err)
		}

		// build slashing tx with witness for spending the staking tx
		slashingTxWithWitness, err := btcDel.BuildSlashingTxWithWitness(bsParams, net, fpSK)
		require.NoError(t, err)

		// assert execution
		btctest.AssertSlashingTxExecution(t, stakingInfo.StakingOutput, slashingTxWithWitness)
	})
}
