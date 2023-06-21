package btcstaking_test

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/btcstaking"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

func FuzzGeneratingParsingValidStakingScript(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		stakerPrivKeyBytes := datagen.GenRandomByteArray(r, 32)
		validatorPrivKeyBytes := datagen.GenRandomByteArray(r, 32)
		juryPrivKeyBytes := datagen.GenRandomByteArray(r, 32)
		stakingTime := uint16(r.Intn(math.MaxUint16))

		_, stakerPublicKey := btcec.PrivKeyFromBytes(stakerPrivKeyBytes)
		_, validatorPublicKey := btcec.PrivKeyFromBytes(validatorPrivKeyBytes)
		_, juryPublicKey := btcec.PrivKeyFromBytes(juryPrivKeyBytes)

		sd, _ := btcstaking.NewStakingScriptData(stakerPublicKey, validatorPublicKey, juryPublicKey, stakingTime)

		script, err := sd.BuildStakingScript()
		require.NoError(t, err)
		parsedScript, err := btcstaking.ParseStakingTransactionScript(0, script)
		require.NoError(t, err)

		require.Equal(t, parsedScript.StakingTime, stakingTime)
		require.Equal(t, schnorr.SerializePubKey(stakerPublicKey), schnorr.SerializePubKey(parsedScript.StakerKey))
		require.Equal(t, schnorr.SerializePubKey(validatorPublicKey), schnorr.SerializePubKey(parsedScript.ValidatorKey))
		require.Equal(t, schnorr.SerializePubKey(juryPublicKey), schnorr.SerializePubKey(parsedScript.JuryKey))
	})
}

// Help function to assert the execution of a script engine. Copied from:
// https://github.com/lightningnetwork/lnd/blob/master/input/script_utils_test.go#L24
func assertEngineExecution(t *testing.T, testNum int, valid bool,
	newEngine func() (*txscript.Engine, error)) {

	t.Helper()

	// Get a new VM to execute.
	vm, err := newEngine()
	require.NoError(t, err, "unable to create engine")

	// Execute the VM, only go on to the step-by-step execution if
	// it doesn't validate as expected.
	vmErr := vm.Execute()
	if valid == (vmErr == nil) {
		return
	}

	// Now that the execution didn't match what we expected, fetch a new VM
	// to step through.
	vm, err = newEngine()
	require.NoError(t, err, "unable to create engine")

	// This buffer will trace execution of the Script, dumping out
	// to stdout.
	var debugBuf bytes.Buffer

	done := false
	for !done {
		dis, err := vm.DisasmPC()
		if err != nil {
			t.Fatalf("stepping (%v)\n", err)
		}
		debugBuf.WriteString(fmt.Sprintf("stepping %v\n", dis))

		done, err = vm.Step()
		if err != nil && valid {
			t.Log(debugBuf.String())
			t.Fatalf("spend test case #%v failed, spend "+
				"should be valid: %v", testNum, err)
		} else if err == nil && !valid && done {
			t.Log(debugBuf.String())
			t.Fatalf("spend test case #%v succeed, spend "+
				"should be invalid: %v", testNum, err)
		}

		debugBuf.WriteString(fmt.Sprintf("Stack: %v", vm.GetStack()))
		debugBuf.WriteString(fmt.Sprintf("AltStack: %v", vm.GetAltStack()))
	}

	// If we get to this point the unexpected case was not reached
	// during step execution, which happens for some checks, like
	// the clean-stack rule.
	validity := "invalid"
	if valid {
		validity = "valid"
	}

	t.Log(debugBuf.String())
	t.Fatalf("%v spend test case #%v execution ended with: %v", validity, testNum, vmErr)
}

func TestStakingScriptExecutionSingleStaker(t *testing.T) {
	const (
		stakingValue      = btcutil.Amount(2 * 10e8)
		stakingTimeBlocks = 5
	)

	r := rand.New(rand.NewSource(time.Now().Unix()))

	stakerPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	validatorPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	juryPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	txid, err := chainhash.NewHash(datagen.GenRandomByteArray(r, 32))
	require.NoError(t, err)

	stakingOut := &wire.OutPoint{
		Hash:  *txid,
		Index: 0,
	}

	stakingOutput, stakingScript, err := btcstaking.BuildStakingOutput(
		stakerPrivKey.PubKey(),
		validatorPrivKey.PubKey(),
		juryPrivKey.PubKey(),
		stakingTimeBlocks,
		stakingValue,
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	spendStakeTx := wire.NewMsgTx(2)

	spendStakeTx.AddTxIn(wire.NewTxIn(stakingOut, nil, nil))

	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			Value:    1 * 10e8,
		},
	)

	// to spend tx as staker, we need to set the sequence number to be >= stakingTimeBlocks
	spendStakeTx.TxIn[0].Sequence = stakingTimeBlocks

	witness, err := btcstaking.BuildWitnessToSpendStakingOutput(
		spendStakeTx,
		stakingOutput,
		stakingScript,
		stakerPrivKey,
	)

	require.NoError(t, err)

	spendStakeTx.TxIn[0].Witness = witness

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingOutput.PkScript, stakingOutput.Value,
	)

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingOutput.Value,
			prevOutputFetcher,
		)
	}
	assertEngineExecution(t, 0, true, newEngine)
}

func TestStakingScriptExecutionMulitSig(t *testing.T) {
	const (
		stakingValue      = btcutil.Amount(2 * 10e8)
		stakingTimeBlocks = 5
	)

	r := rand.New(rand.NewSource(time.Now().Unix()))

	stakerPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	validatorPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	juryPrivKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	txid, err := chainhash.NewHash(datagen.GenRandomByteArray(r, 32))
	require.NoError(t, err)

	stakingOut := &wire.OutPoint{
		Hash:  *txid,
		Index: 0,
	}

	stakingOutput, stakingScript, err := btcstaking.BuildStakingOutput(
		stakerPrivKey.PubKey(),
		validatorPrivKey.PubKey(),
		juryPrivKey.PubKey(),
		stakingTimeBlocks,
		stakingValue,
		&chaincfg.MainNetParams,
	)

	require.NoError(t, err)

	spendStakeTx := wire.NewMsgTx(2)

	spendStakeTx.AddTxIn(wire.NewTxIn(stakingOut, nil, nil))

	spendStakeTx.AddTxOut(
		&wire.TxOut{
			PkScript: []byte("doesn't matter"),
			Value:    1 * 10e8,
		},
	)

	witnessStaker, err := btcstaking.BuildWitnessToSpendStakingOutput(
		spendStakeTx,
		stakingOutput,
		stakingScript,
		stakerPrivKey,
	)
	require.NoError(t, err)

	witnessValidator, err := btcstaking.BuildWitnessToSpendStakingOutput(
		spendStakeTx,
		stakingOutput,
		stakingScript,
		validatorPrivKey,
	)

	require.NoError(t, err)

	witnessJury, err := btcstaking.BuildWitnessToSpendStakingOutput(
		spendStakeTx,
		stakingOutput,
		stakingScript,
		juryPrivKey,
	)

	require.NoError(t, err)

	// To Construct valid witness, for multisig case we need:
	// - jury signature - witnessJury[0]
	// - validator signature - witnessValidator[0]
	// - staker signature - witnessStaker[0]
	// - empty signature - which is just an empty byte array which signals we are going to use multisig.
	// 	 This must be signagure on top of the stack.
	// - whole script - witnessStaker[1] (any other wittness[1] will work as well)
	// - control block - witnessStaker[2] (any other wittness[2] will work as well)
	spendStakeTx.TxIn[0].Witness = [][]byte{
		witnessJury[0], witnessValidator[0], witnessStaker[0], []byte{}, witnessStaker[1], witnessStaker[2],
	}

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(
		stakingOutput.PkScript, stakingOutput.Value,
	)

	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			stakingOutput.PkScript,
			spendStakeTx, 0, txscript.StandardVerifyFlags, nil,
			txscript.NewTxSigHashes(spendStakeTx, prevOutputFetcher), stakingOutput.Value,
			prevOutputFetcher,
		)
	}
	assertEngineExecution(t, 0, true, newEngine)
}
