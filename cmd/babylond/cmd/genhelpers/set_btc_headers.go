package genhelpers

import (
	"fmt"
	"os"

	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	btclighttypes "github.com/babylonchain/babylon/x/btclightclient/types"

	"github.com/spf13/cobra"
)

// CmdSetBtcHeaders CLI sets bitcoin headers into the genesis state.
func CmdSetBtcHeaders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-btc-headers [path/to/btc_headers.json]",
		Short: "Set the BTC headers from the given json file into the genesis.json",
		Long: `Reads BTC Headers structures from the given json file and update the genesis.json file
in place to include the btc headers in the btcstaking module's genesis state.
Duplicated BTC headers are not allowed and it will prompt an error.
`,
		Example: `babylond gen-helpers set-btc-headers path/to/btc_headers.json
Possible content of 'btc_headers.json' is
{
	"btc_headers": [
		{
			"header": "0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a45068653ffff7f2002000000",
			"hash": "683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6",
			"height": "0",
			"work": "2"
		},
		{
			"header": "00000020f67ad7695d9b662a72ff3d8edbbb2de0bfa67b13974bb9910d116d5cbd863e682259b3c6351788d4456e8fbb738dd51fd6aea615f63b10f634021f930c9c34be02150466ffff7f2000000000",
			"hash": "6fceca20e50018d9f54632b791144eef4b3f9cb2ced9fa2702c376857a550e03",
			"height": "1",
			"work": "4"
		}
	]
}
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			config := server.GetServerContextFromCmd(cmd).Config
			config.SetRoot(clientCtx.HomeDir)

			inputBtcHeaders, err := getBtcLightGenStateFromFile(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			btclightGenState := btclighttypes.GenesisStateFromAppState(clientCtx.Codec, appState)

			genStateBtcHeaderByHash := make(map[string]struct{}, 0)
			for _, btcHeader := range btclightGenState.BtcHeaders {
				key := btcHeader.Hash.MarshalHex()
				if _, ok := genStateBtcHeaderByHash[key]; ok {
					return fmt.Errorf("bad genesis state, there is more than one btc header with the same hash %s", key)
				}
				genStateBtcHeaderByHash[key] = struct{}{}
			}

			newBtcHeaders := make([]*btclighttypes.BTCHeaderInfo, 0, len(inputBtcHeaders.BtcHeaders))
			for _, btcHeader := range inputBtcHeaders.BtcHeaders {
				if err := btcHeader.Validate(); err != nil {
					return fmt.Errorf("failed to validate basic btc header: %w", err)
				}

				key := btcHeader.Hash.MarshalHex()
				if _, ok := genStateBtcHeaderByHash[key]; ok {
					return fmt.Errorf("error: btc header: %+v\nwas already set on genesis, or contains the same hash %s than another btc header", btcHeader, key)
				}
				genStateBtcHeaderByHash[key] = struct{}{}
				newBtcHeaders = append(newBtcHeaders, btcHeader)
			}
			btclightGenState.BtcHeaders = append(btclightGenState.BtcHeaders, newBtcHeaders...)

			if err := btclightGenState.Validate(); err != nil {
				return err
			}

			err = replaceModOnGenesis(clientCtx.Codec, genDoc, appState, btclighttypes.ModuleName, &btclightGenState)
			if err != nil {
				return err
			}

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	return cmd
}

func getBtcLightGenStateFromFile(cdc codec.Codec, inputFilePath string) (*btclighttypes.GenesisState, error) {
	if !cmtos.FileExists(inputFilePath) {
		return nil, fmt.Errorf("input file %s does not exists", inputFilePath)
	}

	bz, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, err
	}

	var genState btclighttypes.GenesisState
	err = cdc.UnmarshalJSON(bz, &genState)
	if err != nil {
		return nil, err
	}

	return &genState, nil
}
