package cli

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
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
		Use:   "withdraw-reward [type]",
		Short: "withdraw reward of the stakeholder behind the transaction submitter in a given type (one of {submitter, reporter, finality_provider, btc_delegation})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgWithdrawReward{
				Type:    args[0],
				Address: clientCtx.FromAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
