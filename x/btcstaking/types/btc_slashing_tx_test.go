package types_test

import (
	"bytes"
	"math/rand"
	"sort"
	"testing"

	sdkmath "cosmossdk.io/math"
	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

func FuzzSlashingTxWithWitness(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)

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

		// TODO(restaking): test restaking
		// numRestakedFPs := int(datagen.RandomInt(r, 10) + 1)
		// fpIdx := int(datagen.RandomInt(r, numRestakedFPs))
		numRestakedFPs := 5
		fpIdx := 1
		fpSKs, fpPKs, err := datagen.GenRandomBTCKeyPairs(r, numRestakedFPs)
		require.NoError(t, err)
		fpSK, fpPK := *fpSKs[fpIdx], *fpPKs[fpIdx]

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

		// get covenant Schnorr signatures
		covenantSigs, err := datagen.GenCovenantAdaptorSigs(
			covenantSKs,
			fpPKs,
			stakingMsgTx,
			slashingPkScriptPath,
			slashingTx,
		)
		require.NoError(t, err)
		covSigs, err := types.GetOrderedCovenantSignatures(fpIdx, covenantSigs, &bsParams)
		require.NoError(t, err)

		// get slashed finality provider's signature and its position in the witness
		sortedFPPKs := sortBTCPKs(fpPKs)
		fpIdxInWitness := 0
		found := false
		for i, pk := range sortedFPPKs {
			if pk.IsEqual(&fpPK) {
				fpIdxInWitness = i
				found = true
				break
			}
		}
		require.True(t, found)

		// create slashing tx with witness
		slashingMsgTxWithWitness, err := slashingTx.BuildSlashingTxWithWitness(&fpSK, fpIdxInWitness, numRestakedFPs, stakingMsgTx, 0, delSig, covSigs, slashingSpendInfo)
		require.NoError(t, err)

		// verify slashing tx with witness
		btctest.AssertSlashingTxExecution(t, testStakingInfo.StakingInfo.StakingOutput, slashingMsgTxWithWitness)
	})
}

func sortBTCPKs(keys []*btcec.PublicKey) []*btcec.PublicKey {
	sortedPKs := make([]*btcec.PublicKey, len(keys))
	copy(sortedPKs, keys)
	sort.SliceStable(sortedPKs, func(i, j int) bool {
		keyIBytes := schnorr.SerializePubKey(sortedPKs[i])
		keyJBytes := schnorr.SerializePubKey(sortedPKs[j])
		return bytes.Compare(keyIBytes, keyJBytes) == 1
	})
	return sortedPKs
}
