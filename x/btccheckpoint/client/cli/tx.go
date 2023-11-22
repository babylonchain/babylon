package cli

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
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

	cmd.AddCommand(CmdTxInsertSpvProofs())

	return cmd
}

// TODO this api is not super friendly i.e it is not easy to provide hex encoded
// proto serialized blobs. It would be good to have version which takes some
// other format like json or maybe path to file
func CmdTxInsertSpvProofs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insert-proofs [proof-hex-string] [proof-hex-string]",
		Short: "submit proof bytes",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proof1, err := types.NewSpvProofFromHexBytes(clientCtx.Codec, args[0])

			if err != nil {
				return err
			}

			proof2, err := types.NewSpvProofFromHexBytes(clientCtx.Codec, args[1])

			if err != nil {
				return err
			}

			msg := &types.MsgInsertBTCSpvProof{
				Submitter: clientCtx.GetFromAddress().String(),
				Proofs:    []*types.BTCSpvProof{proof1, proof2},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
