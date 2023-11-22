package types_test

import (
	sdkmath "cosmossdk.io/math"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/btcstaking"
	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
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
		changeAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		// Generate a slashing rate in the range [0.1, 0.50] i.e., 10-50%.
		// NOTE - if the rate is higher or lower, it may produce slashing or change outputs
		// with value below the dust threshold, causing test failure.
		// Our goal is not to test failure due to such extreme cases here;
		// this is already covered in FuzzGeneratingValidStakingSlashingTx
		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)
		valSK, valPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		delSK, delPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		covenantSK, covenantPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		// generate staking/slashing tx
		testStakingInfo := datagen.GenBTCStakingSlashingTx(
			r,
			t,
			net,
			delSK,
			[]*btcec.PublicKey{valPK},
			[]*btcec.PublicKey{covenantPK},
			1,
			stakingTimeBlocks,
			stakingValue,
			slashingAddress.String(), changeAddress.String(),
			slashingRate,
		)

		slashingTx := testStakingInfo.SlashingTx
		stakingMsgTx := testStakingInfo.StakingTx
		stakingPkScript := testStakingInfo.StakingInfo.StakingOutput.PkScript

		slashingScriptInfo, err := testStakingInfo.StakingInfo.SlashingPathSpendInfo()
		require.NoError(t, err)
		slashingScript := slashingScriptInfo.RevealedLeaf.Script

		// sign slashing tx
		valSig, err := slashingTx.Sign(stakingMsgTx, 0, slashingScript, valSK, net)
		require.NoError(t, err)
		delSig, err := slashingTx.Sign(stakingMsgTx, 0, slashingScript, delSK, net)
		require.NoError(t, err)
		covenantSig, err := slashingTx.Sign(stakingMsgTx, 0, slashingScript, covenantSK, net)
		require.NoError(t, err)

		// verify signatures first
		err = slashingTx.VerifySignature(stakingPkScript, stakingValue, slashingScript, valPK, valSig)
		require.NoError(t, err)
		err = slashingTx.VerifySignature(stakingPkScript, stakingValue, slashingScript, delPK, delSig)
		require.NoError(t, err)
		err = slashingTx.VerifySignature(stakingPkScript, stakingValue, slashingScript, covenantPK, covenantSig)
		require.NoError(t, err)

		stakerSigBytes := delSig.MustMarshal()
		validatorSigBytes := valSig.MustMarshal()
		covSigBytes := covenantSig.MustMarshal()

		// TODO: use comittee
		witness, err := btcstaking.CreateBabylonWitness(
			[][]byte{
				covSigBytes,
				validatorSigBytes,
				stakerSigBytes,
			},
			slashingScriptInfo,
		)
		require.NoError(t, err)

		slashingMsgTxWithWitness, err := slashingTx.ToMsgTx()
		require.NoError(t, err)

		slashingMsgTxWithWitness.TxIn[0].Witness = witness
		// verify slashing tx with witness
		prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
			stakingPkScript, stakingValue,
		)
		newEngine := func() (*txscript.Engine, error) {
			return txscript.NewEngine(
				stakingPkScript,
				slashingMsgTxWithWitness, 0, txscript.StandardVerifyFlags, nil,
				txscript.NewTxSigHashes(slashingMsgTxWithWitness, prevOutputFetcher), stakingValue,
				prevOutputFetcher,
			)
		}
		btctest.AssertEngineExecution(t, 0, true, newEngine)
	})
}
