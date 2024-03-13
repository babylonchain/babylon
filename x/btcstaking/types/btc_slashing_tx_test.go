package types_test

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

func FuzzSlashingTxWithWitness(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		var (
			stakingValue      = int64(2 * 10e8)
			stakingTimeBlocks = uint16(5)
			net               = &chaincfg.SimNetParams
		)

		// slashing address and key paris
		slashingAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)

		// TODO: test restaking
		fpSK, fpPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		numRestakedFPs := 1
		fpIdx := 0

		delSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

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
			[]*btcec.PublicKey{fpPK}, // restaking
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

		// sign slashing tx
		delSig, err := slashingTx.Sign(stakingMsgTx, 0, slashingPkScriptPath, delSK)
		require.NoError(t, err)

		covenantSigs, err := datagen.GenCovenantAdaptorSigs(
			covenantSKs,
			[]*btcec.PublicKey{fpPK}, // restaking
			stakingMsgTx,
			slashingPkScriptPath,
			slashingTx,
		)
		require.NoError(t, err)

		covSigs, err := types.GetOrderedCovenantSignatures(0, covenantSigs, &bsParams)
		require.NoError(t, err)

		// create slashing tx with witness
		slashingMsgTxWithWitness, err := slashingTx.BuildSlashingTxWithWitness(fpSK, fpIdx, numRestakedFPs, stakingMsgTx, 0, delSig, covSigs, slashingSpendInfo)
		require.NoError(t, err)

		// verify slashing tx with witness
		btctest.AssertSlashingTxExecution(t, testStakingInfo.StakingInfo.StakingOutput, slashingMsgTxWithWitness)
	})
}
