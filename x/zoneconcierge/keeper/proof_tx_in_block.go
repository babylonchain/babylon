package keeper

import (
	"crypto/sha256"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func (k Keeper) ProveTxInBlock(ctx sdk.Context, txHash []byte) (*tmproto.TxProof, error) {
	if len(txHash) != sha256.Size {
		return nil, fmt.Errorf("invalid txHash length: expected: %d, actual: %d", sha256.Size, len(txHash))
	}

	// query the tx with inclusion proof
	resTx, err := k.tmClient.Tx(ctx, txHash, true)
	if err != nil {
		return nil, err
	}

	txProof := resTx.Proof.ToProto()
	return &txProof, nil
}

func VerifyTxInBlock(txHash []byte, proof *tmproto.TxProof) error {
	txProof, err := tmtypes.TxProofFromProto(*proof)
	if err != nil {
		return err
	}

	return txProof.Proof.Verify(txProof.RootHash, txHash)
}
