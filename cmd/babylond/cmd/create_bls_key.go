package cmd

import (
	"errors"
	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	"path/filepath"
)

var (
	flagAccountAddress = "account-address"
)

func createBlsKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-bls-key [account-address]",
		Short: "Create a pair of BLS keys for a validator",
		Long: `create-bls will create a pair of BLS keys that are used to
send BLS signatures for checkpointing.

BLS keys are stored along with other validator keys in priv_validator_key.json,
which should exist before running the command.

Example:
	babylond create-bls-key --account-address bbn1f5tnl46mk4dfp4nx3n2vnrvyw2h2ydz6ykhk3r --home ./
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
			accAddr, err := cmd.Flags().GetString(flagAccountAddress)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(accAddr)
			if err != nil {
				return err
			}

			return CreateBlsKey(homeDir, addr)
		},
	}

	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "The node home directory")
	cmd.Flags().String(flagAccountAddress, "", "The account address associated with the validator")

	return cmd
}

func CreateBlsKey(home string, addr sdk.AccAddress) error {
	nodeCfg := tmconfig.DefaultConfig()
	keyPath := filepath.Join(home, nodeCfg.PrivValidatorKeyFile())
	statePath := filepath.Join(home, nodeCfg.PrivValidatorStateFile())
	if !tmos.FileExists(keyPath) {
		return errors.New("validator key file does not exist")
	}
	pv := privval.LoadWrappedFilePV(keyPath, statePath)
	wrappedPV := privval.NewWrappedFilePV(pv.GetValPrivKey(), bls12381.GenPrivKey(), keyPath, statePath)
	wrappedPV.SetAccAddress(addr)

	return nil
}
