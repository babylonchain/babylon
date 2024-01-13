package cli

import (
	"fmt"
	"strconv"

	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group incentive queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryRewardGauges(),
		CmdQueryBTCStakingGauge(),
		CmdQueryBTCTimestampingGauge(),
	)

	return cmd
}

func CmdQueryRewardGauges() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reward-gauges [address]",
		Short: "shows reward gauges of a given stakeholder address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryRewardGaugesRequest{
				Address: args[0],
			}
			res, err := queryClient.RewardGauges(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryBTCStakingGauge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-staking-gauge [height]",
		Short: "shows BTC staking gauge of a given height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryBTCStakingGaugeRequest{
				Height: height,
			}
			res, err := queryClient.BTCStakingGauge(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryBTCTimestampingGauge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-timestamping-gauge [epoch]",
		Short: "shows BTC timestamping gauge of a given epoch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			epoch, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryBTCTimestampingGaugeRequest{
				EpochNum: epoch,
			}
			res, err := queryClient.BTCTimestampingGauge(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
