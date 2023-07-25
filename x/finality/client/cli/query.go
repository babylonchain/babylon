package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
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
			finalized, err := cmd.Flags().GetBool("finalized")
			if err != nil {
				return err
			}

			res, err := queryClient.ListBlocks(cmd.Context(), &types.QueryListBlocksRequest{
				Finalized:  finalized,
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
	cmd.Flags().Bool("finalized", false, "return finalized or non-finalized blocks")

	return cmd
}
