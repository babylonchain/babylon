package genhelpers

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// CmdGenHelpers helpers for manipulating the genesis file.
func CmdGenHelpers(validator genutiltypes.MessageValidator) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "gen-helpers",
		Short:                      "Useful commands for creating the genesis state",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdCreateBls(),
		CmdAddBls(validator),
		CmdSetFp(),
	)

	return cmd
}
