package cmd

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/privval"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
)

func GenBlsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-genesis-bls",
		Short: "Create genesis BLS key file for the validator",
		Long: strings.TrimSpace(`genbls will create a BLS key file that consists of
{address, bls_pub_key, pop, pub_key} where pop is the proof-of-possession that proves
the ownership of bls_pub_key which is bonded with pub_key.

The pre-conditions of running the generate-genesis-bls-key are the existence of the keyring,
and the existence of priv_validator_key.json which contains the validator private key.


Example:
$ babylond genbls --home ./
`),

		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)

			nodeCfg := tmconfig.DefaultConfig()
			keyPath := filepath.Join(homeDir, nodeCfg.PrivValidatorKeyFile())
			statePath := filepath.Join(homeDir, nodeCfg.PrivValidatorStateFile())
			if !tmos.FileExists(keyPath) {
				return errors.New("validator key file does not exist")
			}

			wrappedPV := privval.LoadWrappedFilePV(keyPath, statePath)

			outputFileName, err := wrappedPV.ExportGenBls(filepath.Dir(keyPath))
			if err != nil {
				return err
			}

			cmd.PrintErrf("Genesis BLS keys written to %q\n", outputFileName)
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "The node home directory")

	return cmd
}
