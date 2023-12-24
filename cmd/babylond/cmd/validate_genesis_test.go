package cmd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"

	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"

	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/cmd/babylond/cmd"
)

func TestCheckCorrespondence(t *testing.T) {
	homePath := t.TempDir()
	// generate valid genesis doc
	bbn, appState := generateTestGenesisState(t, homePath, 2)
	clientCtx := client.Context{}.WithCodec(bbn.AppCodec()).WithTxConfig(bbn.TxConfig())

	// Copy the appState into a new struct
	bz, err := json.Marshal(appState)
	require.NoError(t, err)
	var mismatchedAppState map[string]json.RawMessage
	err = json.Unmarshal(bz, &mismatchedAppState)
	require.NoError(t, err)

	genutilGenesisState := genutiltypes.GetGenesisStateFromAppState(clientCtx.Codec, mismatchedAppState)
	checkpointingGenesisState := checkpointingtypes.GetGenesisStateFromAppState(clientCtx.Codec, mismatchedAppState)

	// generate mismatched genesis doc by deleting one item from gentx and genKeys in different positions
	genutilGenesisState.GenTxs = genutilGenesisState.GenTxs[:1]
	checkpointingGenesisState.GenesisKeys = checkpointingGenesisState.GenesisKeys[1:]

	// Update the for the genutil module with the invalid data
	genTxsBz, err := clientCtx.Codec.MarshalJSON(genutilGenesisState)
	require.NoError(t, err)
	mismatchedAppState[genutiltypes.ModuleName] = genTxsBz

	// Update the for the checkpointing module with the invalid data
	genKeysBz, err := clientCtx.Codec.MarshalJSON(&checkpointingGenesisState)
	require.NoError(t, err)
	mismatchedAppState[checkpointingtypes.ModuleName] = genKeysBz

	testCases := []struct {
		name     string
		appState map[string]json.RawMessage
		expErr   bool
	}{
		{
			"valid genesis gentx and BLS key pair",
			appState,
			false,
		},
		{
			"mismatched genesis state",
			mismatchedAppState,
			true,
		},
	}

	gentxModule := bbn.BasicModuleManager[genutiltypes.ModuleName].(genutil.AppModuleBasic)

	for _, tc := range testCases {
		err = cmd.CheckCorrespondence(clientCtx, tc.appState, gentxModule.GenTxValidator)
		if tc.expErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func generateTestGenesisState(t *testing.T, home string, n int) (*app.BabylonApp, map[string]json.RawMessage) {
	logger := log.NewNopLogger()
	cfg, _ := genutiltest.CreateDefaultCometConfig(home)

	signer, err := app.SetupTestPrivSigner()
	require.NoError(t, err)
	bbn := app.NewBabylonAppWithCustomOptions(t, false, signer, app.SetupOptions{
		Logger:             logger,
		DB:                 dbm.NewMemDB(),
		InvCheckPeriod:     0,
		SkipUpgradeHeights: map[int64]bool{},
		AppOpts:            app.EmptyAppOptions{},
	})

	_ = genutiltest.ExecInitCmd(bbn.BasicModuleManager, home, bbn.AppCodec())

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.
		WithCodec(bbn.AppCodec()).
		WithHomeDir(home).
		WithTxConfig(bbn.TxConfig())

	ctx := context.Background()
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	testnetCmd := cmd.TestnetCmd(bbn.BasicModuleManager, banktypes.GenesisBalancesIterator{})
	testnetCmd.SetArgs([]string{
		fmt.Sprintf("--%s=test", flags.FlagKeyringBackend),
		fmt.Sprintf("--v=%v", n),
		fmt.Sprintf("--output-dir=%s", home),
	})
	_ = testnetCmd.ExecuteContext(ctx)

	genFile := cfg.GenesisFile()
	genState, _, _ := genutiltypes.GenesisStateFromGenFile(genFile)
	return bbn, genState
}
