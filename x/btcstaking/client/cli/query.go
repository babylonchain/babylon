package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"

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
	cmd.AddCommand(CmdBTCDelegations())
	cmd.AddCommand(CmdBTCValidatorsAtHeight())
	cmd.AddCommand(CmdBTCValidatorPowerAtHeight())
	cmd.AddCommand(CmdActivatedHeight())
	cmd.AddCommand(CmdBTCValidatorDelegations())

	return cmd
}

func CmdBTCValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-validators",
		Short: "retrieve all BTC validators",
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
	flags.AddPaginationFlagsToCmd(cmd, "btc-validators")

	return cmd
}

func CmdBTCDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-delegations [status]",
		Short: "retrieve all BTC delegations under the given status (pending, active, unbonding, unbonded, any)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			status, err := types.NewBTCDelegationStatusFromString(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.BTCDelegations(cmd.Context(), &types.QueryBTCDelegationsRequest{
				Status:     status,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "btc-delegations")

	return cmd
}

func CmdBTCValidatorPowerAtHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-validator-power-at-height [val_btc_pk_hex] [height]",
		Short: "get the voting power of a given BTC validator at a given height",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			res, err := queryClient.BTCValidatorPowerAtHeight(cmd.Context(), &types.QueryBTCValidatorPowerAtHeightRequest{
				ValBtcPkHex: args[0],
				Height:      height,
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

func CmdActivatedHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activated-height",
		Short: "get activated height, i.e., the first height where there exists 1 BTC validator with voting power",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ActivatedHeight(cmd.Context(), &types.QueryActivatedHeightRequest{})
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
		Short: "retrieve all BTC validators at a given babylon height",
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

			res, err := queryClient.ActiveBTCValidatorsAtHeight(cmd.Context(), &types.QueryActiveBTCValidatorsAtHeightRequest{
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
	flags.AddPaginationFlagsToCmd(cmd, "btc-validators-at-height")

	return cmd
}

func CmdBTCValidatorDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-validator-delegations [btc_val_pk_hex]",
		Short: "retrieve all delegations under a given BTC validator",
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
	flags.AddPaginationFlagsToCmd(cmd, "btc-validator-delegations")

	return cmd
}
