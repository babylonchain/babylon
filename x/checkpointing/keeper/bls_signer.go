package keeper

import (
	"context"
	"fmt"
	"time"

	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/client/tx"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/types/retry"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

type BlsSigner interface {
	GetAddress() sdk.ValAddress
	SignMsgWithBls(msg []byte) (bls12381.Signature, error)
	GetBlsPubkey() (bls12381.PublicKey, error)
}

// SendBlsSig prepares a BLS signature message and sends it to Comet
func (k Keeper) SendBlsSig(ctx context.Context, epochNum uint64, appHash types.AppHash, valSet epochingtypes.ValidatorSet) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// get self address
	addr := k.blsSigner.GetAddress()

	// check if itself is the validator
	_, _, err := valSet.FindValidatorWithIndex(addr)
	if err != nil {
		// only send the BLS sig when the node itself is a validator, not being a validator is not an error
		return nil
	}

	// get BLS signature by signing
	signBytes := types.GetSignBytes(epochNum, appHash)
	blsSig, err := k.blsSigner.SignMsgWithBls(signBytes)
	if err != nil {
		return err
	}

	// create MsgAddBlsSig message
	msg := types.NewMsgAddBlsSig(k.clientCtx.GetFromAddress(), epochNum, appHash, blsSig, addr)

	// keep sending the message to Comet until success or timeout
	// TODO should read the parameters from config file
	var res *sdk.TxResponse
	err = retry.Do(1*time.Second, 1*time.Minute, func() error {
		res, err = tx.SendMsgToComet(ctx, k.clientCtx, msg)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		sdkCtx.Logger().Error(fmt.Sprintf("Failed to send the BLS sig tx for epoch %v: %v", epochNum, err))
		return err
	}

	sdkCtx.Logger().Info(fmt.Sprintf("Successfully sent BLS-sig tx for epoch %d, tx hash: %s, gas used: %d, gas wanted: %d", epochNum, res.TxHash, res.GasUsed, res.GasWanted))

	return nil
}
