package genhelpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	btcstktypes "github.com/babylonchain/babylon/x/btcstaking/types"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"
)

// CmdSetFp CLI sets finality providers into the genesis state.
func CmdSetFp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-finality-providers [path/to/finality_providers.json]",
		Short: "Set the finality providers from the given json file into the genesis.json",
		Long: `Reads finality providers structures from the given json file and update the genesis.json file
in place to include the finality providers in the btcstaking module's genesis state.
Duplicated finality providers are not allowed and it will prompt an error.
`,
		Example: `babylond gen-helpers set-finality-providers path/to/finality_providers.json
Possible content of 'finality_providers.json' is
{
	"finality_providers": [
		{
			"description": {
				"moniker": "val-fp",
				"identity": "",
				"website": "",
				"security_contact": "",
				"details": ""
			},
			"commission": "0.050000000000000000",
			"babylon_pk": {
				"key": "A6FTeYaKkvi8cOXSmVF+2n+q58WV/qffI99YucsvrGsk"
			},
			"btc_pk": "625057f9828753569c478b1aa826205395222bcdbefafe9653dfdc638d18eadc",
			"pop": {
				"btc_sig_type": "BIP340",
				"babylon_sig": "/RJyVqBEm/PDkQQTDgdPj84MywTu6LUVXDWUlPBgDGo1qsU/Rg+kJhBRwxCEKs8PYmoADcp4FiH2vVyAqRG2UA==",
				"btc_sig": "OaLj+w2loH0O88cT2IEdUoV4Swx+RWagk632+mKs5wVRBM63D6uwW+lCLIjxVBfOYQRFzT2D9943k10Whw5a7Q=="
			},
			"slashed_babylon_height": "0",
			"slashed_btc_height": "0"
		}
	]
}
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			config := server.GetServerContextFromCmd(cmd).Config
			config.SetRoot(clientCtx.HomeDir)

			inputFps, err := getBtcStakingGenStateFromFile(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			btcstkGenState := btcstktypes.GenesisStateFromAppState(clientCtx.Codec, appState)

			genStateFpsByBtcPk, err := mapFinalityProvidersByBtcPk(btcstkGenState.FinalityProviders)
			if err != nil {
				return fmt.Errorf("bad gen state: %w", err)
			}

			newFps := make([]*btcstktypes.FinalityProvider, 0, len(inputFps.FinalityProviders))
			for _, fp := range inputFps.FinalityProviders {
				if err := fp.ValidateBasic(); err != nil {
					return fmt.Errorf("failed to validate basic finality provider: %w", err)
				}

				key := fp.BtcPk.MarshalHex()
				if _, ok := genStateFpsByBtcPk[key]; ok {
					return fmt.Errorf("error: finality provider: %+v\nwas already set on genesis, or contains the same BtcPk %s than another finality provider", fp, key)
				}

				// sets the fp to the genstate to avoid having 2 fps with same btc pk in the input
				genStateFpsByBtcPk[key] = struct{}{}
				newFps = append(newFps, fp)
			}
			btcstkGenState.FinalityProviders = append(btcstkGenState.FinalityProviders, newFps...)

			err = replaceModOnGenesis(clientCtx.Codec, genDoc, appState, btcstktypes.ModuleName, &btcstkGenState)
			if err != nil {
				return err
			}

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	return cmd
}

func replaceModOnGenesis(
	cdc codec.Codec,
	genDoc *genutiltypes.AppGenesis,
	appState map[string]json.RawMessage,
	modname string,
	modGenState proto.Message,
) error {
	newModGenState, err := cdc.MarshalJSON(modGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal %s genesis state: %w", modname, err)
	}
	appState[modname] = newModGenState

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}
	genDoc.AppState = appStateJSON
	return nil
}

func getBtcStakingGenStateFromFile(cdc codec.Codec, inputFilePath string) (*btcstktypes.GenesisState, error) {
	filePath := filepath.Clean(inputFilePath)
	if !cmtos.FileExists(filePath) {
		return nil, fmt.Errorf("input file %s does not exists", filePath)
	}

	bz, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var genState btcstktypes.GenesisState
	err = cdc.UnmarshalJSON(bz, &genState)
	if err != nil {
		return nil, err
	}

	return &genState, nil
}

func mapFinalityProvidersByBtcPk(fps []*btcstktypes.FinalityProvider) (map[string]struct{}, error) {
	genStateFpsByBtcPk := make(map[string]struct{}, 0)
	for _, fpGen := range fps {
		key := fpGen.BtcPk.MarshalHex()
		if _, ok := genStateFpsByBtcPk[key]; ok {
			return nil, fmt.Errorf("there is more than one finality provider with the same btc key %s", key)
		}
		genStateFpsByBtcPk[key] = struct{}{}
	}
	return genStateFpsByBtcPk, nil
}
