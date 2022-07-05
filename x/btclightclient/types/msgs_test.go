package types_test

import (
	"bytes"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/blockchain"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/big"
	"math/rand"
	"testing"
)

func FuzzMsgInsertHeader(f *testing.F) {
	addressStr := "from________________"
	addressBytes := []byte(addressStr)
	maxDifficulty, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

	f.Add(addressBytes, types.DefaultBaseHeaderHex, int64(17))
	f.Fuzz(func(t *testing.T, addressBytes []byte, headerHex string, seed int64) {
		seedInput := types.DefaultBaseHeaderHex == headerHex
		errorKind := 0

		// Populate proper data
		rand.Seed(seed)
		if len(addressBytes) == 0 || len(addressBytes) >= 256 {
			addressBytes = genRandomByteArray(1 + uint64(rand.Intn(257)))
		}
		if !validHex(headerHex, bbl.BTCHeaderLen) {
			headerHex = genRandomHexStr(bbl.BTCHeaderLen)
		}

		// Even if the data has the proper structure, some checks still fail
		var signer sdk.AccAddress
		err := signer.Unmarshal(addressBytes)
		if err != nil {
			// Invalid address
			t.Skip()
		}
		tmpHeader, err := bbl.NewBTCHeaderBytesFromHex(headerHex)
		if err != nil {
			t.Skip()
		}
		tmpBtcdHeader, err := tmpHeader.ToBlockHeader()
		if err != nil {
			// Cannot be converted to a block header
			t.Skip()
		}

		// Change the btcd header and convert back to hex
		if !seedInput {
			errorKind = rand.Intn(3)
			switch errorKind {
			case 0:
				// Valid input
				// Set the work bits to the pow limit
				tmpBtcdHeader.Bits = blockchain.BigToCompact(maxDifficulty)
			case 1:
				// Zero PoW
				tmpBtcdHeader.Bits = blockchain.BigToCompact(big.NewInt(0))
			case 2:
				// Negative PoW
				tmpBtcdHeader.Bits = blockchain.BigToCompact(big.NewInt(-1))
			}
		}
		header := bbl.NewBTCHeaderBytesFromBlockHeader(tmpBtcdHeader)
		headerHex, err = header.MarshalHex()
		if err != nil {
			// btcd header can't be marshalled to hex
			t.Skip()
		}

		// Check the message creation
		msgInsertHeader, err := types.NewMsgInsertHeader(signer, headerHex)
		if err != nil {
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

		// Validate the message
		err = msgInsertHeader.ValidateHeader(maxDifficulty)
		if err != nil && errorKind == 0 {
			t.Errorf("Valid message %s failed with %s", headerHex, err)
		}
		if err == nil && errorKind != 0 {
			t.Errorf("Invalid message did not fail")
		}
	})

}
