package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
)

func (k Keeper) ProveTxInBlock(ctx context.Context, txHash []byte) (*tmproto.TxProof, error) {
	if len(txHash) != sha256.Size {
		return nil, fmt.Errorf("invalid txHash length: expected: %d, actual: %d", sha256.Size, len(txHash))
	}

	// query the tx with inclusion proof
	resTx, err := k.cmtClient.Tx(ctx, txHash, true)
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
