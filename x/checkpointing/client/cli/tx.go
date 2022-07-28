package cli

import (
	"fmt"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"strconv"

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
