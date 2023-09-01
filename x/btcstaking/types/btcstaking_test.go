package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

func FuzzStakingTx(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		net := &chaincfg.SimNetParams

		stakerSK, stakerPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		_, validatorPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		_, juryPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		stakingTimeBlocks := uint16(5)
		stakingValue := int64(2 * 10e8)
		slashingAddr, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		stakingTx, _, err := datagen.GenBTCStakingSlashingTx(r, net, stakerSK, validatorPK, juryPK, stakingTimeBlocks, stakingValue, slashingAddr.String())
		require.NoError(t, err)

		err = stakingTx.ValidateBasic()
		require.NoError(t, err)

		// extract staking script and staked value
		stakingOutputInfo, err := stakingTx.GetBabylonOutputInfo(&chaincfg.SimNetParams)
		require.NoError(t, err)
		// NOTE: given that PK derived from SK has 2 possibilities on a curve, we can only compare x value but not y value
		require.Equal(t, stakingOutputInfo.StakingScriptData.StakerKey.SerializeCompressed()[1:], stakerPK.SerializeCompressed()[1:])
		require.Equal(t, stakingOutputInfo.StakingScriptData.ValidatorKey.SerializeCompressed()[1:], validatorPK.SerializeCompressed()[1:])
		require.Equal(t, stakingOutputInfo.StakingScriptData.JuryKey.SerializeCompressed()[1:], juryPK.SerializeCompressed()[1:])
		require.Equal(t, stakingOutputInfo.StakingScriptData.StakingTime, stakingTimeBlocks)
		require.Equal(t, int64(stakingOutputInfo.StakingAmount), stakingValue)
	})
}

func FuzzBTCDelegation(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		btcDel := &types.BTCDelegation{}
		// randomise voting power
		btcDel.TotalSat = datagen.RandomInt(r, 100000)

		// randomise jury sig
		hasJurySig := datagen.RandomInt(r, 2) == 0
		if hasJurySig {
			jurySig := bbn.BIP340Signature([]byte{1, 2, 3})
			btcDel.JurySig = &jurySig
		}

		// randomise start height and end height
		btcDel.StartHeight = datagen.RandomInt(r, 100)
		btcDel.EndHeight = btcDel.StartHeight + datagen.RandomInt(r, 100)

		// randomise BTC tip and w
		btcHeight := btcDel.StartHeight + datagen.RandomInt(r, 50)
		w := datagen.RandomInt(r, 50)

		// test expected voting power
		hasVotingPower := hasJurySig && btcDel.StartHeight <= btcHeight && btcHeight+w <= btcDel.EndHeight
		actualVotingPower := btcDel.VotingPower(btcHeight, w)
		if hasVotingPower {
			require.Equal(t, btcDel.TotalSat, actualVotingPower)
		} else {
			require.Equal(t, uint64(0), actualVotingPower)
		}
	})
}
