package types

import (
	fmt "fmt"
	"math/big"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/x/btccheckpoint/btcutils"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// Ensure that MsgInsertBTCSpvProof implements all functions of the Msg interface
	_ sdk.Msg = (*MsgInsertBTCSpvProof)(nil)
)

// Parse and Validate transactions which should contain OP_RETURN data.
// OP_RETURN bytes are not validated in any way. It is up to the caller attach
// semantic meaning and validity to those bytes.
// Returned ParsedProofs are in same order as raw proofs
// TODO explore possibility of validating that output in second tx is payed by
// input in the first tx
func ParseTwoProofs(submitter sdk.AccAddress, proofs []*BTCSpvProof, powLimit *big.Int) (*RawCheckpointSubmission, error) {
	// Expecting as many proofs as many parts our checkpoint is composed of
	if len(proofs) != txformat.NumberOfParts {
		return nil, fmt.Errorf("expected at exactly valid op return transactions")
	}

	var parsedProofs []*btcutils.ParsedProof

	for _, proof := range proofs {
		parsedProof, e :=
			btcutils.ParseProof(
				proof.BtcTransaction,
				proof.BtcTransactionIndex,
				proof.MerkleNodes,
				proof.ConfirmingBtcHeader,
				powLimit,
			)

		if e != nil {
			return nil, e
		}

		parsedProofs = append(parsedProofs, parsedProof)
	}

	var checkpointData [][]byte

	for i, proof := range parsedProofs {
		// TODO tag should be taken from configuration
		data, err := txformat.GetCheckpointData(txformat.MainTag, txformat.CurrentVersion, uint8(i), proof.OpReturnData)
		if err != nil {
			return nil, err
		}
		checkpointData = append(checkpointData, data)
	}

	// at this point we know we have two correctly formated babylon op return transacitons
	// we need to check if parts match
	fullTxData, err := txformat.ConnectParts(txformat.CurrentVersion, checkpointData[0], checkpointData[1])

	if err != nil {
		return nil, err
	}

	sub := NewRawCheckpointSubmission(submitter, *parsedProofs[0], *parsedProofs[1], fullTxData)

	return &sub, nil
}

func (m *MsgInsertBTCSpvProof) ValidateBasic() error {
	address, err := sdk.AccAddressFromBech32(m.Submitter)

	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid submitter address: %s", err)
	}

	// TODO get powLimit from some config
	// result of parsed proof is not needed, drop it
	// whole parsing stuff is stateless
	_, err = ParseTwoProofs(address, m.Proofs, btcchaincfg.MainNetParams.PowLimit)

	if err != nil {
		return err
	}

	return nil
}

func (m *MsgInsertBTCSpvProof) GetSigners() []sdk.AccAddress {
	// cosmos-sdk modules usually ignore possible error here, we panic for the sake
	// of informing something terrible had happend

	submitter, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		// Panic, since the GetSigners method is called after ValidateBasic
		// which performs the same check.
		panic(err)
	}

	return []sdk.AccAddress{submitter}
}
