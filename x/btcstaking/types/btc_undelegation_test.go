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
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

func FuzzBTCUndelegation_SlashingTx(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		net := &chaincfg.SimNetParams

		delSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		fpSK, fpPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		fpPKList := []*btcec.PublicKey{fpPK}

		// (3, 5) covenant committee
		covenantSKs, covenantPKs, err := datagen.GenRandomBTCKeyPairs(r, 5)
		require.NoError(t, err)
		covenantQuorum := uint32(3)

		stakingTimeBlocks := uint16(5)
		stakingValue := int64(2 * 10e8)
		slashingAddress, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)

		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)
		unbondingTime := uint16(100) + 1
		unbondingValue := stakingValue - 1000
		slashingChangeLockTime := unbondingTime

		// construct the BTC delegation with everything
		btcDel, err := datagen.GenRandomBTCDelegation(
			r,
			t,
			bbn.NewBIP340PKsFromBTCPKs(fpPKList),
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

		stakingTxHash := btcDel.MustGetStakingTxHash()

		testInfo := datagen.GenBTCUnbondingSlashingInfo(
			r,
			t,
			net,
			delSK,
			fpPKList,
			covenantPKs,
			covenantQuorum,
			wire.NewOutPoint(&stakingTxHash, 0),
			unbondingTime,
			unbondingValue,
			slashingAddress.EncodeAddress(),
			slashingRate,
			slashingChangeLockTime,
		)
		require.NoError(t, err)

		// delegator signs the unbonding slashing tx
		delSlashingTxSig, err := testInfo.GenDelSlashingTxSig(delSK)
		require.NoError(t, err)

		unbondingTxBytes, err := bbn.SerializeBTCTx(testInfo.UnbondingTx)
		require.NoError(t, err)

		// spend info of the unbonding slashing tx
		unbondingSlashingSpendInfo, err := testInfo.UnbondingInfo.SlashingPathSpendInfo()
		require.NoError(t, err)
		// covenant signs (using adaptor signature) the slashing tx
		covenantSigs, err := datagen.GenCovenantAdaptorSigs(
			covenantSKs,
			[]*btcec.PublicKey{fpPK},
			testInfo.UnbondingTx,
			unbondingSlashingSpendInfo.GetPkScriptPath(),
			testInfo.SlashingTx,
		)
		require.NoError(t, err)
		btcDel.BtcUndelegation = &types.BTCUndelegation{
			UnbondingTx:              unbondingTxBytes,
			SlashingTx:               testInfo.SlashingTx,
			DelegatorUnbondingSig:    nil, // not relevant here
			DelegatorSlashingSig:     delSlashingTxSig,
			CovenantSlashingSigs:     covenantSigs,
			CovenantUnbondingSigList: nil, // not relevant here
		}

		bsParams := &types.Params{
			CovenantPks:    bbn.NewBIP340PKsFromBTCPKs(covenantPKs),
			CovenantQuorum: covenantQuorum,
		}

		// build slashing tx with witness for spending the unbonding tx
		unbondingSlashingTxWithWitness, err := btcDel.BuildUnbondingSlashingTxWithWitness(bsParams, net, fpSK)
		require.NoError(t, err)

		// assert the execution
		btctest.AssertSlashingTxExecution(t, testInfo.UnbondingInfo.UnbondingOutput, unbondingSlashingTxWithWitness)
	})
}
