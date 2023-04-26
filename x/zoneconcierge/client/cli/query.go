package cli

import (
	"context"
	"fmt"

	// "strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/zoneconcierge/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group zoneconcierge queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdChainInfo())
	cmd.AddCommand(CmdChainsInfo())
	return cmd
}

func CmdChainInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain-info <chain-id>",
		Short: "retrieve the chain info of given chain id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)
			req := types.QueryChainInfoRequest{ChainId: args[0]}
			resp, err := queryClient.ChainInfo(context.Background(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdChainsInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chains-info <chain-ids>",
		Short: "retrieve the chain info of given chain ids",
		Args:  cobra.RangeArgs(1, 5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)
			req := types.QueryChainsInfoRequest{ChainIds: args}
			resp, err := queryClient.ChainsInfo(context.Background(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
