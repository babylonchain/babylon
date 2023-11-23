package types_test

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

func FuzzStakingTx(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		net := &chaincfg.SimNetParams

		stakerSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		_, validatorPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		_, covenantPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		stakingTimeBlocks := uint16(5)
		stakingValue := int64(2 * 10e8)
		slashingAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		changeAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)
		testInfo := datagen.GenBTCStakingSlashingTx(
			r,
			t,
			net,
			stakerSK,
			[]*btcec.PublicKey{validatorPK},
			[]*btcec.PublicKey{covenantPK},
			1,
			stakingTimeBlocks,
			stakingValue,
			slashingAddress.EncodeAddress(), changeAddress.EncodeAddress(),
			slashingRate,
		)
		require.NoError(t, err)
		require.Equal(t, testInfo.StakingInfo.StakingOutput.Value, stakingValue)
	})
}

func FuzzBTCDelegation(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		btcDel := &types.BTCDelegation{}
		// randomise voting power
		btcDel.TotalSat = datagen.RandomInt(r, 100000)

		// randomise covenant sig
		hasCovenantSig := datagen.RandomInt(r, 2) == 0
		if hasCovenantSig {
			covenantSig := bbn.BIP340Signature([]byte{1, 2, 3})
			btcDel.CovenantSig = &covenantSig
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
