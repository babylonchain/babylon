package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/babylonchain/babylon/x/checkpointing/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group headeroracle queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(CmdRawCheckpoint())
	cmd.AddCommand(CmdRawCheckpointList())
	cmd.AddCommand(CmdRawCheckpoints())

	return cmd
}

// CmdRawCheckpointList defines the cobra command to query raw checkpoints by status
func CmdRawCheckpointList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raw-checkpoint-list [status]",
		Short: "retrieve the checkpoints by status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			status, exists := types.CheckpointStatus_value[args[0]]
			if !exists {
				return errors.New("invalid status")
			}

			params := types.NewQueryRawCheckpointListRequest(pageReq, types.CheckpointStatus(status))
			res, err := queryClient.RawCheckpointList(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdRawCheckpoint defines the cobra command to query the raw checkpoint by epoch number
func CmdRawCheckpoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raw-checkpoint [epoch_number]",
		Short: "retrieve the checkpoint by epoch number",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			epoch_num, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			params := types.NewQueryRawCheckpointRequest(epoch_num)
			res, err := queryClient.RawCheckpoint(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdRawCheckpoints defines the cobra command to query the raw checkpoints
func CmdRawCheckpoints() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raw-checkpoints",
		Short: "retrieve the checkpoints for a epoch range",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.RawCheckpoints(context.Background(), &types.QueryRawCheckpointsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "raw-checkpoints")

	return cmd
}
