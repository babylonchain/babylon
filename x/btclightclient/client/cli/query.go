package cli

import (
	"context"
	"fmt"

	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group btclightclient queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdHashes())
	cmd.AddCommand(CmdContains())
	cmd.AddCommand(CmdMainChain())
	cmd.AddCommand(CmdTip())
	cmd.AddCommand(CmdBaseHeader())
	cmd.AddCommand(CmdHeaderDepth())

	return cmd
}

func CmdHashes() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hashes",
		Short: "retrieve the hashes maintained by this module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := types.NewQueryHashesRequest(pageReq)
			res, err := queryClient.Hashes(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "hashes")

	return cmd
}

func CmdContains() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contains [hex-hash]",
		Short: "check whether the module maintains a hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params, err := types.NewQueryContainsRequest(args[0])
			if err != nil {
				return err
			}
			res, err := queryClient.Contains(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdMainChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "main-chain",
		Short: "retrieve the canonical chain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := types.NewQueryMainChainRequest(pageReq)
			res, err := queryClient.MainChain(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "main-chain")

	return cmd
}

func CmdTip() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tip",
		Short: "retrieve tip of the bitcoin blockchain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := types.NewQueryTipRequest()
			res, err := queryClient.Tip(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdBaseHeader() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base-header",
		Short: "retrieve base header of the bitcoin blockchain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := types.NewQueryBaseHeaderRequest()
			res, err := queryClient.BaseHeader(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdHeaderDepth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "header-depth [hex-hash]",
		Short: "check main chain depth of the header with the given hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			depthRequest, err := types.NewQueryHeaderDepthRequest(args[0])
			if err != nil {
				return err
			}
			res, err := queryClient.HeaderDepth(context.Background(), depthRequest)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
