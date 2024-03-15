package bitcoin

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

// Help function to assert the execution of a script engine. Copied from:
// https://github.com/lightningnetwork/lnd/blob/master/input/script_utils_test.go#L24
func AssertEngineExecution(t *testing.T, testNum int, valid bool,
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
			t.Fatalf("spend test case #%v failed at (%v), spend "+
				"should be valid: %v", testNum, dis, err)
		} else if err == nil && !valid && done {
			t.Log(debugBuf.String())
			t.Fatalf("spend test case #%v succeed at (%v), spend "+
				"should be invalid: %v", testNum, dis, err)
		}

		debugBuf.WriteString(fmt.Sprintf("Stack: %v\n", vm.GetStack()))
		debugBuf.WriteString(fmt.Sprintf("AltStack: %v\n\n", vm.GetAltStack()))
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

// AssertSlashingTxExecution asserts that the given tx has a valid witness for spending the given funding output
func AssertSlashingTxExecution(t *testing.T, fundingOutput *wire.TxOut, txWithWitness *wire.MsgTx) {
	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(fundingOutput.PkScript, fundingOutput.Value)
	newEngine := func() (*txscript.Engine, error) {
		return txscript.NewEngine(
			fundingOutput.PkScript,
			txWithWitness,
			0,
			txscript.StandardVerifyFlags,
			nil,
			txscript.NewTxSigHashes(txWithWitness, prevOutputFetcher),
			fundingOutput.Value,
			prevOutputFetcher,
		)
	}
	AssertEngineExecution(t, 0, true, newEngine)
}
