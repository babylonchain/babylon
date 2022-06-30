package types_test

import (
	"bytes"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
)

func FuzzMsgInsertHeader(f *testing.F) {
	addressStr := "from________________"
	addressBytes := []byte(addressStr)

	f.Add(addressBytes, types.DefaultBaseHeaderHex)
	f.Fuzz(func(t *testing.T, addressBytes []byte, headerHex string) {
		msgCreationShouldFail := false
		validateShouldFail := false

		var signer sdk.AccAddress
		err := signer.Unmarshal(addressBytes)
		if err != nil {
			// Invalid address
			t.Skip()
		}

		header, err := bbl.NewBTCHeaderBytesFromHex(headerHex)
		if err != nil {
			// Invalid header, later this should fail
			msgCreationShouldFail = true
		}

		btcdHeader, err := header.ToBlockHeader()
		if err != nil {
			// Cannot be converted to a block header
			t.Skip()
		}

		// Create a block to check proof of work
		msgBlock := &wire.MsgBlock{Header: *btcdHeader}
		block := btcutil.NewBlock(msgBlock)
		// TODO: get the parameter from a configuration file
		if blockchain.CheckProofOfWork(block, btcchaincfg.MainNetParams.PowLimit) != nil {
			// There's an invalid work parameter
			validateShouldFail = true
		}

		if len(addressBytes) == 0 || len(addressBytes) >= 256 {
			// signers should be between 1-256 bytes
			validateShouldFail = true
		}

		msgInsertHeader, err := types.NewMsgInsertHeader(signer, headerHex)
		if err != nil {
			if msgCreationShouldFail {
				// All good, this should have failed
				t.Skip()
			}
			t.Errorf("Valid parameters led to error")
		}

		if msgInsertHeader == nil {
			t.Errorf("nil returned")
		}

		if msgInsertHeader.Header == nil {
			t.Errorf("nil header")
		}

		if bytes.Compare(header, *(msgInsertHeader.Header)) != 0 {
			t.Errorf("Expected header bytes %s got %s", header, *(msgInsertHeader.Header))
		}

		err = msgInsertHeader.ValidateBasic()
		if err != nil {
			if validateShouldFail {
				t.Skip()
			}
			t.Errorf("Validation for valid message failed with err %s", err)
		}
	})

}
