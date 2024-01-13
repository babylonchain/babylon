package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cosmossdk.io/core/address"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	appparams "github.com/babylonchain/babylon/app/params"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cosmoscli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/babylonchain/babylon/x/checkpointing/types"
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

	cmd.AddCommand(CmdWrappedCreateValidator(authcodec.NewBech32Codec(appparams.Bech32PrefixValAddr)))

	return cmd
}

func CmdWrappedCreateValidator(valAddrCodec address.Codec) *cobra.Command {
	cmd := cosmoscli.NewCreateValidatorCmd(valAddrCodec)
	cmd.Long = strings.TrimSpace(`create-validator will create a new validator initialized
with a self-delegation to it using the BLS key generated for the validator (e.g., via babylond create-bls-key).

This command creates a MsgWrappedCreateValidator message which is a wrapper of cosmos-sdk's
create validator with a pair of BLS key. The BLS key should exist in priv_validator_key.json
before running the command (e.g., via babylond create-bls-key).`)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
		if err != nil {
			return err
		}

		val, err := parseAndValidateValidatorJSON(clientCtx.Codec, args[0])
		if err != nil {
			return err
		}

		txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

		txf, msg, err := buildWrappedCreateValidatorMsg(clientCtx, txf, cmd.Flags(), val, valAddrCodec)
		if err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
	}
	// HACK: test cases need to setup the path where the priv validator BLS key is going to be set
	// so we redefine the FlagHome here. Since we can't import `app` due to a cyclic dependency,
	// we have to duplicate the definition here.
	// If this changes, the `DefaultHomeDir` flag at `app/app.go` needs to change as well.
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	defaultNodeHome := filepath.Join(userHomeDir, ".babylond")
	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The node home directory")

	return cmd
}
