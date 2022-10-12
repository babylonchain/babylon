package keeper

import (
	"fmt"
	"github.com/babylonchain/babylon/client/tx"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BlsSigner interface {
	GetAddress() sdk.ValAddress
	SignMsgWithBls(msg []byte) (bls12381.Signature, error)
	GetBlsPubkey() (bls12381.PublicKey, error)
}

// SendBlsSig prepares a BLS signature message and sends it to Tendermint
func (k Keeper) SendBlsSig(ctx sdk.Context, epochNum uint64, lch types.LastCommitHash) (*sdk.TxResponse, error) {
	// get self address
	curValSet := k.GetValidatorSet(ctx, epochNum)
	addr := k.blsSigner.GetAddress()

	// check if itself is the validator
	_, _, err := curValSet.FindValidatorWithIndex(addr)
	if err != nil {
		// only send the BLS sig when the node itself is a validator, not being a validator is not an error
		return nil, nil
	}

	// get BLS signature by signing
	signBytes := append(sdk.Uint64ToBigEndian(epochNum), lch...)
	blsSig, err := k.blsSigner.SignMsgWithBls(signBytes)
	if err != nil {
		return nil, err
	}

	// create MsgAddBlsSig message
	msg := types.NewMsgAddBlsSig(epochNum, lch, blsSig, addr)

	// send the message to Tendermint
	res, err := tx.SendMsgToTendermint(k.clientCtx, msg)
	// TODO this is retriable
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Failed to send the BLS sig tx for epoch %v: %v", epochNum, err))
		return nil, err
	}

	return res, nil
}
