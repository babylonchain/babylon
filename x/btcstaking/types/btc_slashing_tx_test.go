package types_test

import (
	"math/rand"
	"testing"

	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		slashingRate := sdk.NewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)
		valSK, valPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		delSK, delPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		covenantSK, covenantPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		// generate staking/slashing tx
		stakingTx, slashingTx, err := datagen.GenBTCStakingSlashingTx(
			r,
			net,
			delSK,
			[]*btcec.PublicKey{valPK},
			[]*btcec.PublicKey{covenantPK},
			stakingTimeBlocks,
			stakingValue,
			slashingAddress.String(), changeAddress.String(),
			slashingRate,
		)
		require.NoError(t, err)
		stakingOutInfo, err := stakingTx.GetBabylonOutputInfo(net)
		require.NoError(t, err)
		stakingPkScript := stakingOutInfo.StakingPkScript
		stakingMsgTx, err := stakingTx.ToMsgTx()
		require.NoError(t, err)

		// sign slashing tx
		valSig, err := slashingTx.Sign(stakingMsgTx, stakingTx.Script, valSK, net)
		require.NoError(t, err)
		delSig, err := slashingTx.Sign(stakingMsgTx, stakingTx.Script, delSK, net)
		require.NoError(t, err)
		covenantSig, err := slashingTx.Sign(stakingMsgTx, stakingTx.Script, covenantSK, net)
		require.NoError(t, err)

		// verify signatures first
		err = slashingTx.VerifySignature(stakingPkScript, stakingValue, stakingTx.Script, valPK, valSig)
		require.NoError(t, err)
		err = slashingTx.VerifySignature(stakingPkScript, stakingValue, stakingTx.Script, delPK, delSig)
		require.NoError(t, err)
		err = slashingTx.VerifySignature(stakingPkScript, stakingValue, stakingTx.Script, covenantPK, covenantSig)
		require.NoError(t, err)

		// build slashing tx with witness
		slashingMsgTxWithWitness, err := slashingTx.ToMsgTxWithWitness(stakingTx, valSig, delSig, covenantSig)
		require.NoError(t, err)

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
