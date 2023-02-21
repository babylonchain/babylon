package cmd_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/server/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tmconfig "github.com/tendermint/tendermint/config"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/tempfile"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/cmd/babylond/cmd"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/testutil/cli"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

// test adding genesis BLS keys without gentx
// error is expected
func Test_AddGenBlsCmdWithoutGentx(t *testing.T) {
	home := t.TempDir()
	logger := log.NewNopLogger()
	tmcfg, err := genutiltest.CreateDefaultTendermintConfig(home)
	require.NoError(t, err)

	appCodec := app.GetEncodingConfig().Marshaler
	err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), tmcfg, logger)
	clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)
	cfg := serverCtx.Config
	cfg.SetRoot(clientCtx.HomeDir)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	genKey := datagen.GenerateGenesisKey()
	jsonBytes, err := tmjson.MarshalIndent(genKey, "", "  ")
	require.NoError(t, err)
	genKeyFileName := filepath.Join(home, fmt.Sprintf("gen-bls-%s.json", genKey.ValidatorAddress))
	err = tempfile.WriteFileAtomic(genKeyFileName, jsonBytes, 0600)
	require.NoError(t, err)
	addGenBlsCmd := cmd.AddGenBlsCmd()
	addGenBlsCmd.SetArgs(
		[]string{genKeyFileName},
	)
	err = addGenBlsCmd.ExecuteContext(ctx)
	require.Error(t, err)
}

// test adding genesis BLS keys with gentx
// error is expected if adding duplicate
func Test_AddGenBlsCmdWithGentx(t *testing.T) {
	cfg := network.DefaultConfig()
	config.SetConfigTemplate(config.DefaultConfigTemplate)
	cfg.NumValidators = 1

	testNetwork, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	defer testNetwork.Cleanup()

	_, err = testNetwork.WaitForHeight(1)
	require.NoError(t, err)

	targetCfg := tmconfig.DefaultConfig()
	targetCfg.SetRoot(filepath.Join(testNetwork.Validators[0].Dir, "simd"))
	targetGenesisFile := targetCfg.GenesisFile()
	targetCtx := testNetwork.Validators[0].ClientCtx
	for i := 0; i < cfg.NumValidators; i++ {
		v := testNetwork.Validators[i]
		// build and create genesis BLS key
		genBlsCmd := cmd.GenBlsCmd()
		nodeCfg := tmconfig.DefaultConfig()
		homeDir := filepath.Join(v.Dir, "simd")
		nodeCfg.SetRoot(homeDir)
		keyPath := nodeCfg.PrivValidatorKeyFile()
		statePath := nodeCfg.PrivValidatorStateFile()
		filePV := privval.GenWrappedFilePV(keyPath, statePath)
		defer filePV.Clean(keyPath, statePath)
		filePV.SetAccAddress(v.Address)
		_, err = cli.ExecTestCLICmd(v.ClientCtx, genBlsCmd, []string{fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir)})
		require.NoError(t, err)
		genKeyFileName := filepath.Join(filepath.Dir(keyPath), fmt.Sprintf("gen-bls-%s.json", v.ValAddress))
		genKey, err := types.LoadGenesisKeyFromFile(genKeyFileName)
		require.NoError(t, err)
		require.NotNil(t, genKey)

		// add genesis BLS key to the target context
		addBlsCmd := cmd.AddGenBlsCmd()
		_, err = cli.ExecTestCLICmd(targetCtx, addBlsCmd, []string{genKeyFileName})
		require.NoError(t, err)
		appState, _, err := genutiltypes.GenesisStateFromGenFile(targetGenesisFile)
		require.NoError(t, err)
		// test duplicate
		_, err = cli.ExecTestCLICmd(targetCtx, addBlsCmd, []string{genKeyFileName})
		require.Error(t, err)

		checkpointingGenState := types.GetGenesisStateFromAppState(v.ClientCtx.Codec, appState)
		require.NotEmpty(t, checkpointingGenState.GenesisKeys)
		gks := checkpointingGenState.GetGenesisKeys()
		require.Equal(t, genKey, gks[i])
	}
}
