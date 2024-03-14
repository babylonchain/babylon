package types_test

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

func FuzzBTCUndelegation_SlashingTx(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		net := &chaincfg.SimNetParams

		delSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		// restaked to a random number of finality providers
		numRestakedFPs := int(datagen.RandomInt(r, 10) + 1)
		fpSKs, fpPKs, err := datagen.GenRandomBTCKeyPairs(r, numRestakedFPs)
		fpBTCPKs := bbn.NewBIP340PKsFromBTCPKs(fpPKs)
		require.NoError(t, err)

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

		// construct the BTC delegation with everything
		btcDel, err := datagen.GenRandomBTCDelegation(
			r,
			t,
			fpBTCPKs,
			delSK,
			covenantSKs,
			covenantQuorum,
			slashingAddress.EncodeAddress(),
			1000,
			uint64(1000+stakingTimeBlocks),
			uint64(stakingValue),
			slashingRate,
			slashingChangeLockTime,
		)
		require.NoError(t, err)

		unbondingInfo, err := btcDel.GetUnbondingInfo(bsParams, net)
		require.NoError(t, err)

		// build slashing tx with witness for spending the unbonding tx
		// a random finality provider gets slashed
		slashedFPIdx := int(datagen.RandomInt(r, numRestakedFPs))
		fpSK := fpSKs[slashedFPIdx]
		unbondingSlashingTxWithWitness, err := btcDel.BuildUnbondingSlashingTxWithWitness(bsParams, net, fpSK)
		require.NoError(t, err)

		// assert the execution
		btctest.AssertSlashingTxExecution(t, unbondingInfo.UnbondingOutput, unbondingSlashingTxWithWitness)
	})
}
