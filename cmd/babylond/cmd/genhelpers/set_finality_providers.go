package genhelpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/spf13/cobra"
)

// CmdSetFp CLI sets finality providers into the genesis state.
func CmdSetFp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-finality-providers [path/to/finality_providers.json]",
		Short: "Set the finality providers from the given json file into the genesis.json",
		Long: `Reads finality providers structures from the given json file and update the genesis.json file
in place to include the finality providers in the btcstaking module's genesis state.
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

			finalityProvidersInputPath := args[0]
			if !cmtos.FileExists(finalityProvidersInputPath) {
				return errors.New("finality providers input file does not exist")
			}

			fpsBz, err := os.ReadFile(finalityProvidersInputPath)
			if err != nil {
				return err
			}

			var inputFps btcstakingtypes.GenesisState
			err = clientCtx.Codec.UnmarshalJSON(fpsBz, &inputFps)
			if err != nil {
				return err
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			btcstkGenState := btcstakingtypes.GenesisStateFromAppState(clientCtx.Codec, appState)

			var newFps []*btcstakingtypes.FinalityProvider
			for _, fp := range inputFps.FinalityProviders {
				if err := fp.ValidateBasic(); err != nil {
					return fmt.Errorf("failed to validate basic finality provider: %w", err)
				}

				for _, fpGen := range btcstkGenState.FinalityProviders {
					if fp.BtcPk.Equals(fpGen.BtcPk) {
						return fmt.Errorf("error: finality provider: %+v\nwas already set on genesis, or contains the same BtcPk %s than another finality provider", fp, fpGen.BtcPk.MarshalHex())
					}
				}

				newFps = append(newFps, fp)
			}
			btcstkGenState.FinalityProviders = append(btcstkGenState.FinalityProviders, newFps...)

			btcstkGenStateWithFps, err := clientCtx.Codec.MarshalJSON(&btcstkGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal btcstaking genesis state: %w", err)
			}
			appState[btcstakingtypes.ModuleName] = btcstkGenStateWithFps

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
