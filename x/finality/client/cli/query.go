package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/babylonchain/babylon/x/finality/types"
)

const (
	flagQueriedBlockStatus = "queried-block-status"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group finality queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdListPublicRandomness())
	cmd.AddCommand(CmdBlock())
	cmd.AddCommand(CmdListBlocks())
	cmd.AddCommand(CmdVotesAtHeight())

	return cmd
}

func CmdVotesAtHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "votes-at-height [height]",
		Short: "retrieve all BTC val pks who voted at requested babylon height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.VotesAtHeight(cmd.Context(), &types.QueryVotesAtHeightRequest{Height: height})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdListPublicRandomness() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-public-randomness [val_btc_pk_hex]",
		Short: "list public randomness committed by a given BTC validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.ListPublicRandomness(cmd.Context(), &types.QueryListPublicRandomnessRequest{
				ValBtcPkHex: args[0],
				Pagination:  pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "list-public-randomness")

	return cmd
}

func CmdListBlocks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-blocks",
		Short: "list blocks at a given status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queriedBlockStatusString, err := cmd.Flags().GetString(flagQueriedBlockStatus)
			if err != nil {
				return err
			}
			queriedBlockStatus, err := types.NewQueriedBlockStatus(queriedBlockStatusString)
			if err != nil {
				return err
			}

			res, err := queryClient.ListBlocks(cmd.Context(), &types.QueryListBlocksRequest{
				Status:     queriedBlockStatus,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "list-blocks")
	cmd.Flags().String(flagQueriedBlockStatus, "Any", "Status of the queried blocks (NonFinalized|Finalized|Any)")

	return cmd
}

func CmdBlock() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block [height]",
		Short: "show the information of the block at a given height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			queriedBlockHeight, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.Block(cmd.Context(), &types.QueryBlockRequest{
				Height: queriedBlockHeight,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
