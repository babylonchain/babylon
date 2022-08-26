package keeper

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/pflag"
	tmconfig "github.com/tendermint/tendermint/config"
)

// SendBlsSig prepares a BLS signature message and sends it to Tendermint
func (k Keeper) SendBlsSig(ctx sdk.Context, epochNum uint64, lch types.LastCommitHash) error {
	// get self address
	curValSet := k.GetValidatorSet(ctx, epochNum)
	conf := tmconfig.DefaultConfig()
	conf.PrivValidatorKeyFile()
	wrappedPV := privval.LoadWrappedFilePV(conf.PrivValidatorKey, conf.PrivValidatorState)
	addr := sdk.ValAddress(wrappedPV.GetAddress())

	// check if itself is the validator
	_, _, err := curValSet.FindValidatorWithIndex(addr)
	if err != nil {
		return err
	}

	// get BLS signature by signing
	blsPrivKey := wrappedPV.GetBlsPrivKey()
	signBytes := append(sdk.Uint64ToBigEndian(epochNum), lch...)
	blsSig := bls12381.Sign(blsPrivKey, signBytes)

	// create MsgAddBlsSig message
	msg := types.NewMsgAddBlsSig(epochNum, lch, blsSig, addr)

	// insert the message into the transaction
	fs := pflag.NewFlagSet("checkpointing", pflag.ContinueOnError)
	err = fs.Set(flags.FlagFrom, k.GetParams(ctx).String())
	if err != nil {
		return err
	}
	err = tx.GenerateOrBroadcastTxCLI(client.Context{}, fs, msg)
	if err != nil {
		return err
	}
	
	return nil
}
