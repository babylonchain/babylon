package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"strconv"

	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group btcstaking queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdBTCValidators())
	cmd.AddCommand(CmdBTCValidatorsAtHeight())
	cmd.AddCommand(CmdBTCValidatorDelegations())

	return cmd
}

func CmdBTCValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-validators",
		Short: "retrieve all btc validators",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.BTCValidators(cmd.Context(), &types.QueryBTCValidatorsRequest{
				Pagination: pageReq,
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

func CmdBTCValidatorsAtHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-validators-at-height [height]",
		Short: "retrieve all btc validators at a given babylon height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.BTCValidatorsAtHeight(cmd.Context(), &types.QueryBTCValidatorsAtHeightRequest{
				Height:     height,
				Pagination: pageReq,
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

func CmdBTCValidatorDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-validator-delegations [btc_val_pk_hex]",
		Short: "retrieve all delegations under a given btc validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.BTCValidatorDelegations(cmd.Context(), &types.QueryBTCValidatorDelegationsRequest{
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

	return cmd
}
