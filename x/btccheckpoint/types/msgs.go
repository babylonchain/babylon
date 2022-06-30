package types

import (
	fmt "fmt"
	"math/big"

	"github.com/babylonchain/babylon/x/btccheckpoint/btcutils"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// Ensure that MsgInsertBTCSpvProof implements all functions of the Msg interface
	_ sdk.Msg = (*MsgInsertBTCSpvProof)(nil)
)

const (
	// two proofs is babylon specific not bitcoin specific, that why it is defined
	// here not in btcutils
	// This could also be a parameter in config. At this time babylon expect 2,
	// OP_RETRUN transactions with valid proofs.
	expectedProofs = 2
)

// Parse and Validate transactions which should contain OP_RETURN data.
// OP_RETURN bytes are not validated in any way. It is up to the caller attach
// semantic meaning and validity to those bytes.
// Returned ParsedProofs are in same order as raw proofs
// TODO explore possibility of validating that output in second tx is payed by
// input in the first tx
func ParseTwoProofs(proofs []*BTCSpvProof, powLimit *big.Int) ([]*btcutils.ParsedProof, error) {
	if len(proofs) != expectedProofs {
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

	return parsedProofs, nil
}

func (m *MsgInsertBTCSpvProof) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Submitter); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid submitter address: %s", err)
	}

	// TODO get powLimit from some config
	// result of parsed proof is not needed, drop it
	// whole parsing stuff is stateless
	_, err := ParseTwoProofs(m.Proofs, btcchaincfg.MainNetParams.PowLimit)

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
