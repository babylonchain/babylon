package keeper

import (
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func (k Keeper) ProveTxInBlock(ctx sdk.Context, txHash []byte) (tmproto.TxProof, error) {
	if len(txHash) != sha256.Size {
		return tmproto.TxProof{}, fmt.Errorf("invalid txHash length: expected: %d, actual: %d", sha256.Size, len(txHash))
	}

	// get the Tendermint client based on client context
	nodeURI := "tcp://localhost:26657"
	tmClient, err := client.NewClientFromNode(nodeURI)
	if err != nil {
		return tmproto.TxProof{}, fmt.Errorf("couldn't get client from nodeURI: %v", err)
	}
	defer tmClient.Stop()

	// query the tx with inclusion proof
	resTx, err := tmClient.Tx(ctx, txHash, true)
	if err != nil {
		return tmproto.TxProof{}, err
	}

	return resTx.Proof.ToProto(), nil
}

func VerifyTxInBlock(txHash []byte, proof tmproto.TxProof) error {
	txProof, err := tmtypes.TxProofFromProto(proof)
	if err != nil {
		return err
	}

	return txProof.Proof.Verify(txProof.RootHash, txHash)
}
