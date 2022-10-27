package cmd_test

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/tempfile"

	"github.com/babylonchain/babylon/app"
	bbncmd "github.com/babylonchain/babylon/cmd/babylond/cmd"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

func FuzzAddGenBlsCmd(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		home := t.TempDir()
		logger := log.NewNopLogger()
		cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
		require.NoError(t, err)

		appCodec := app.MakeTestEncodingConfig().Marshaler
		err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
		require.NoError(t, err)

		serverCtx := server.NewContext(viper.New(), cfg, logger)
		clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)
		config := serverCtx.Config
		config.SetRoot(clientCtx.HomeDir)

		ctx := context.Background()
		ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
		ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

		n := rand.Intn(100)
		genKeys := make([]*types.GenesisKey, n)
		// test adding n genesis BLS keys
		for i := 0; i < n; i++ {
			genKeys[i] = datagen.GenerateGenesisKey()
			jsonBytes, err := tmjson.MarshalIndent(genKeys[i], "", "  ")
			require.NoError(t, err)
			genKeyFileName := filepath.Join(home, fmt.Sprintf("gen-bls-%s.json", genKeys[i].ValidatorAddress))
			err = tempfile.WriteFileAtomic(genKeyFileName, jsonBytes, 0600)
			require.NoError(t, err)
			cmd := bbncmd.AddGenBlsCmd()
			cmd.SetArgs(
				[]string{genKeyFileName},
			)
			err = cmd.ExecuteContext(ctx)
			require.NoError(t, err)

			genFile := config.GenesisFile()
			appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
			require.NoError(t, err)

			checkpointingGenState := types.GetGenesisStateFromAppState(clientCtx.Codec, appState)
			require.NotEmpty(t, checkpointingGenState.GenesisKeys)
			gks := checkpointingGenState.GetGenesisKeys()
			require.Equal(t, genKeys[i], gks[i])
		}

		// test adding a duplicate genesis BLS key
		if n >= 1 {
			cmd2 := bbncmd.AddGenBlsCmd()
			cmd2.SetArgs(
				[]string{filepath.Join(home, fmt.Sprintf("gen-bls-%s.json", genKeys[0].ValidatorAddress))},
			)
			err = cmd2.ExecuteContext(ctx)
			require.Error(t, err)
		}
	})
}
