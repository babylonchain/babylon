package genhelpers

import (
	"encoding/json"
	"errors"
	"fmt"

	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"

	"github.com/babylonchain/babylon/x/checkpointing/types"
)

// CmdAddBls CLI adds the BLS key file with proof of possesion into the genesis state.
func CmdAddBls(validator genutiltypes.MessageValidator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-bls [genesis_bls_file]",
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

			// load genesis BLS key
			genKeyFilePath := args[0]
			if !cmtos.FileExists(genKeyFilePath) {
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

			// load genesis state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			checkpointingGenState := types.GetGenesisStateFromAppState(clientCtx.Codec, appState)

			// check duplication
			gks := checkpointingGenState.GetGenesisKeys()
			for _, gk := range gks {
				if gk.ValidatorAddress == genKey.ValidatorAddress {
					return errors.New("validator address already exists")
				}
			}

			// check correspondence of genesis transactions
			// each genesis BLS key should have a corresponding
			// genesis transaction
			genTxState := genutiltypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)
			foundInGenTx := false
			for _, genTx := range genTxState.GenTxs {
				tx, err := genutiltypes.ValidateAndGetGenTx(genTx, clientCtx.TxConfig.TxJSONDecoder(), validator)
				if err != nil {
					return err
				}
				msgs := tx.GetMsgs()
				if len(msgs) == 0 {
					return errors.New("invalid genesis transaction")
				}
				msgCreateValidator := msgs[0].(*stakingtypes.MsgCreateValidator)
				if msgCreateValidator.ValidatorAddress == genKey.ValidatorAddress {
					foundInGenTx = true
				}
			}
			if !foundInGenTx {
				return errors.New("corresponding genesis tx is not found, add genesis tx with the same validator address first")
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
