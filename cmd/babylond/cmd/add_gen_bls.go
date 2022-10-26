package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	tmos "github.com/tendermint/tendermint/libs/os"

	"github.com/babylonchain/babylon/x/checkpointing/types"
)

func AddGenBlsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-bls [genesis_bls_file]",
		Short: "Add a genesis BLS key to genesis.json",
		Long: `Add a genesis BLS key per validator and update the pregenesis file in place to include their
BLS keys in the checkpointing module's genesis state.'
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genKeyFilePath := args[0]
			if !tmos.FileExists(genKeyFilePath) {
				return errors.New("genesis BLS key file does not exist")
			}

			genKey, err := types.LoadGenesisKeyFromFile(genKeyFilePath)
			if err != nil {
				return err
			}
			err = genKey.Validate()
			if err != nil {
				return err
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			checkpointingGenState := types.GetGenesisStateFromAppState(clientCtx.Codec, appState)
			gks := checkpointingGenState.GetGenesisKeys()
			// deduplicate
			for _, gk := range gks {
				if gk.ValidatorAddress == genKey.ValidatorAddress {
					return errors.New("validator address already exist")
				}
			}
			gks = append(gks, genKey)
			checkpointingGenState.GenesisKeys = gks

			checkpointingGenStateBz, err := clientCtx.Codec.MarshalJSON(&checkpointingGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal checkpointing genesis state: %w", err)
			}

			appState[types.ModuleName] = checkpointingGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	return cmd
}
