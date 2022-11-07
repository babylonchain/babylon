package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/babylonchain/babylon/x/checkpointing/types"
)

const chainUpgradeGuide = "https://github.com/cosmos/cosmos-sdk/blob/a51aa517c46c70df04a06f586c67fb765e45322a/UPGRADING.md"

// ValidateGenesisCmd takes a genesis file, and makes sure that it is valid.
// 1. genesis state of each module should be valid according to each module's
// validation rule
// 2. each genesis BLS key or gentx should have a corresponding gentx or genesis
// BLS key
// modified based on "https://github.com/cosmos/cosmos-sdk/blob/6d32debf1aca4b7f1ed1429d87be1d02c315f02d/x/genutil/client/cli/validate_genesis.go"
func ValidateGenesisCmd(mbm module.BasicManager) *cobra.Command {
	return &cobra.Command{
		Use:   "validate-genesis [file]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "validates the genesis file at the default location or at the location passed as an arg",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx := client.GetClientContextFromCmd(cmd)

			cdc := clientCtx.Codec

			// Load default if passed no args, otherwise load passed file
			var genesis string
			if len(args) == 0 {
				genesis = serverCtx.Config.GenesisFile()
			} else {
				genesis = args[0]
			}

			genDoc, err := validateGenDoc(genesis)
			if err != nil {
				return err
			}

			var genState map[string]json.RawMessage
			if err = json.Unmarshal(genDoc.AppState, &genState); err != nil {
				return fmt.Errorf("error unmarshalling genesis doc %s: %s", genesis, err.Error())
			}

			if err = mbm.ValidateGenesis(cdc, clientCtx.TxConfig, genState); err != nil {
				return fmt.Errorf("error validating genesis file %s: %s", genesis, err.Error())
			}

			if err = CheckCorrespondence(clientCtx, genState); err != nil {
				return fmt.Errorf("error validating genesis file correspondence %s: %s", genesis, err.Error())
			}

			fmt.Printf("File at %s is a valid genesis file\n", genesis)
			return nil
		},
	}
}

// validateGenDoc reads a genesis file and validates that it is a correct
// Tendermint GenesisDoc. This function does not do any cosmos-related
// validation.
func validateGenDoc(importGenesisFile string) (*tmtypes.GenesisDoc, error) {
	genDoc, err := tmtypes.GenesisDocFromFile(importGenesisFile)
	if err != nil {
		return nil, fmt.Errorf("%s. Make sure that"+
			" you have correctly migrated all Tendermint consensus params, please see the"+
			" chain migration guide at %s for more info",
			err.Error(), chainUpgradeGuide,
		)
	}

	return genDoc, nil
}

// CheckCorrespondence checks that each genesis tx/BLS key should have one
// corresponding BLS key/genesis tx
func CheckCorrespondence(ctx client.Context, genesis map[string]json.RawMessage) error {
	checkpointingGenState := types.GetGenesisStateFromAppState(ctx.Codec, genesis)
	gks := checkpointingGenState.GetGenesisKeys()
	genTxState := genutiltypes.GetGenesisStateFromAppState(ctx.Codec, genesis)
	addresses := make(map[string]struct{}, 0)
	// ensure no duplicate BLS keys
	for _, gk := range gks {
		addresses[gk.ValidatorAddress] = struct{}{}
	}
	if len(addresses) != len(gks) {
		return errors.New("duplicate genesis BLS keys")
	}
	// ensure the number of BLS keys and gentxs are the same so that we
	// don't need to do reverse checking
	if len(addresses) != len(genTxState.GenTxs) {
		return errors.New("genesis txs and genesis BLS keys do not match")
	}
	// ensure every gentx has a match with BLS key by address
	for _, genTx := range genTxState.GenTxs {
		tx, err := genutiltypes.ValidateAndGetGenTx(genTx, ctx.TxConfig.TxJSONDecoder())
		if err != nil {
			return err
		}
		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return errors.New("invalid genesis transaction")
		}
		msgCreateValidator := msgs[0].(*stakingtypes.MsgCreateValidator)
		if _, exists := addresses[msgCreateValidator.ValidatorAddress]; !exists {
			return errors.New("cannot find a corresponding BLS key for a genesis tx")
		}
	}

	return nil
}
