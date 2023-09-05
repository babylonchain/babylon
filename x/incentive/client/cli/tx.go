package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
)

var (
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())
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

	cmd.AddCommand(
		NewWithdrawRewardCmd(),
	)

	return cmd
}

func NewWithdrawRewardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-reward [type] [address]",
		Short: "withdraw reward of a given stakeholder in a given type (one of {submitter, reporter, btc_validator, btc_delegation})",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgWithdrawReward{
				Signer:  clientCtx.FromAddress.String(),
				Type:    args[0],
				Address: args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
