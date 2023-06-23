package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

func FuzzStakingTx(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

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
		stakingTx, _, err := datagen.GenBTCStakingSlashingTx(r, stakerSK, validatorPK, juryPK, stakingTimeBlocks, stakingValue, slashingAddr)
		require.NoError(t, err)

		err = stakingTx.ValidateBasic()
		require.NoError(t, err)

		// extract staking script and staked value
		stakingOutputInfo, err := stakingTx.GetStakingOutputInfo(&chaincfg.SimNetParams)
		require.NoError(t, err)
		// NOTE: given that PK derived from SK has 2 possibilities on a curve, we can only compare x value but not y value
		require.Equal(t, stakingOutputInfo.StakingScriptData.StakerKey.SerializeCompressed()[1:], stakerPK.SerializeCompressed()[1:])
		require.Equal(t, stakingOutputInfo.StakingScriptData.ValidatorKey.SerializeCompressed()[1:], validatorPK.SerializeCompressed()[1:])
		require.Equal(t, stakingOutputInfo.StakingScriptData.JuryKey.SerializeCompressed()[1:], juryPK.SerializeCompressed()[1:])
		require.Equal(t, stakingOutputInfo.StakingScriptData.StakingTime, stakingTimeBlocks)
		require.Equal(t, int64(stakingOutputInfo.StakingAmount), stakingValue)
	})
}
