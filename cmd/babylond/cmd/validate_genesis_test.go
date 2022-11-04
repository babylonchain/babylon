package cmd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/cmd/babylond/cmd"
	"github.com/babylonchain/babylon/x/checkpointing/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestCheckCorrespondence(t *testing.T) {
	homePath := t.TempDir()
	encodingCft := app.MakeTestEncodingConfig()
	clientCtx := client.Context{}.WithCodec(encodingCft.Marshaler).WithTxConfig(encodingCft.TxConfig)

	// generate valid genesis doc
	validGenState, genDoc := generateTestGenesisState(homePath, 2)
	validGenDocJSON, err := tmjson.MarshalIndent(genDoc, "", "  ")
	require.NoError(t, err)

	// generate mismatched genesis doc by deleting one item from gentx and genKeys in different positions
	gentxs := genutiltypes.GetGenesisStateFromAppState(clientCtx.Codec, validGenState)
	genKeys := types.GetGenesisStateFromAppState(clientCtx.Codec, validGenState)
	gentxs.GenTxs = gentxs.GenTxs[:1]
	genKeys.GenesisKeys = genKeys.GenesisKeys[1:]
	genTxsBz, err := clientCtx.Codec.MarshalJSON(gentxs)
	require.NoError(t, err)
	genKeysBz, err := clientCtx.Codec.MarshalJSON(&genKeys)
	require.NoError(t, err)
	validGenState[genutiltypes.ModuleName] = genTxsBz
	validGenState[types.ModuleName] = genKeysBz
	misMatchedGenStateBz, err := json.Marshal(validGenState)
	require.NoError(t, err)
	genDoc.AppState = misMatchedGenStateBz
	misMatchedGenDocJSON, err := tmjson.MarshalIndent(genDoc, "", "  ")
	require.NoError(t, err)

	testCases := []struct {
		name    string
		genesis []byte
		expErr  bool
	}{
		{
			"valid genesis gentx and BLS key pair",
			validGenDocJSON,
			false,
		},
		{
			"mismatched genesis state",
			misMatchedGenDocJSON,
			true,
		},
	}

	for _, tc := range testCases {
		genDoc, err := tmtypes.GenesisDocFromJSON(tc.genesis)
		require.NoError(t, err)
		require.NotEmpty(t, genDoc)
		genesisState, err := genutiltypes.GenesisStateFromGenDoc(*genDoc)
		require.NoError(t, err)
		require.NotEmpty(t, genesisState)
		err = cmd.CheckCorrespondence(clientCtx, genesisState)
		if tc.expErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func generateTestGenesisState(home string, n int) (map[string]json.RawMessage, *tmtypes.GenesisDoc) {
	encodingConfig := app.MakeTestEncodingConfig()
	logger := log.NewNopLogger()
	cfg, _ := genutiltest.CreateDefaultTendermintConfig(home)

	_ = genutiltest.ExecInitCmd(app.ModuleBasics, home, encodingConfig.Marshaler)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithHomeDir(home).
		WithTxConfig(encodingConfig.TxConfig)

	ctx := context.Background()
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	testnetCmd := cmd.TestnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{})
	testnetCmd.SetArgs([]string{
		fmt.Sprintf("--%s=test", flags.FlagKeyringBackend),
		fmt.Sprintf("--v=%v", n),
		fmt.Sprintf("--output-dir=%s", home),
	})
	_ = testnetCmd.ExecuteContext(ctx)

	genFile := cfg.GenesisFile()
	appState, gendoc, _ := genutiltypes.GenesisStateFromGenFile(genFile)
	return appState, gendoc
}
