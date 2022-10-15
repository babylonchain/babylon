package cli

import (
	"fmt"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoscli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/spf13/cobra"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

const (
	flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	listSeparator              = ","
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdTxAddBlsSig())
	cmd.AddCommand(CmdWrappedCreateValidator())

	return cmd
}

func CmdTxAddBlsSig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit [epoch_number] [last_commit_hash] [bls_sig] [signer address]",
		Short: "submit a BLS signature",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			epoch_num, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			lch, err := types.NewLastCommitHashFromHex(args[1])
			if err != nil {
				return err
			}

			blsSig, err := bls12381.NewBLSSigFromHex(args[2])
			if err != nil {
				return err
			}

			addr, err := sdk.ValAddressFromBech32(args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgAddBlsSig(epoch_num, lch, blsSig, addr)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdWrappedCreateValidator() *cobra.Command {
	cmd := cosmoscli.NewCreateValidatorCmd()
	cmd.Long = strings.TrimSpace(
		fmt.Sprintf(`create-validator will create a new validator initialized
with a self-delegation to it using the BLS key generated for the validator (e.g., via babylond create-bls-key).

This command creates a MsgWrappedCreateValidator message which is a wrapper of cosmos-sdk's
create validator with a pair of BLS key. The BLS key should exist in priv_validator_key.json
before running the command (e.g., via babylond create-bls-key).`))
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).
			WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
		txf, msg, err := buildWrappedCreateValidatorMsg(clientCtx, txf, cmd.Flags())
		if err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
	}

	cmd.Flags().String(flags.FlagHome, "", "The node home directory")
	_ = cmd.MarkFlagRequired(flags.FlagHome)

	return cmd
}
