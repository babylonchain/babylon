package types_test

import (
	"math/rand"
	"testing"

	btctest "github.com/babylonchain/babylon/testutil/bitcoin"
	"github.com/babylonchain/babylon/testutil/datagen"
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
		slashingAddr, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		valSK, valPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		delSK, delPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		jurySK, juryPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		// generate staking/slashing tx
		stakingTx, slashingTx, err := datagen.GenBTCStakingSlashingTx(r, net, delSK, valPK, juryPK, stakingTimeBlocks, stakingValue, slashingAddr.String())
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
		jurySig, err := slashingTx.Sign(stakingMsgTx, stakingTx.Script, jurySK, net)
		require.NoError(t, err)

		// verify signatures first
		err = slashingTx.VerifySignature(stakingPkScript, stakingValue, stakingTx.Script, valPK, valSig)
		require.NoError(t, err)
		err = slashingTx.VerifySignature(stakingPkScript, stakingValue, stakingTx.Script, delPK, delSig)
		require.NoError(t, err)
		err = slashingTx.VerifySignature(stakingPkScript, stakingValue, stakingTx.Script, juryPK, jurySig)
		require.NoError(t, err)

		// build slashing tx with witness
		slashingMsgTxWithWitness, err := slashingTx.ToMsgTxWithWitness(stakingTx, valSig, delSig, jurySig)
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
