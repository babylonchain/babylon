package genhelpers_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"

	"cosmossdk.io/log"
	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/cmd/babylond/cmd/genhelpers"
	"github.com/babylonchain/babylon/testutil/datagen"
	btcstktypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cometbft/cometbft/libs/tempfile"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/stretchr/testify/require"
)

func FuzzCmdSetFp(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		qntFps := int(datagen.RandomInt(r, 10)) + 1
		fpsToAdd := make([]*btcstktypes.FinalityProvider, qntFps)
		for i := 0; i < qntFps; i++ {
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)
			fpsToAdd[i] = fp
		}

		home := t.TempDir()
		signer, err := app.SetupTestPrivSigner()
		require.NoError(t, err)

		bbn := app.NewBabylonAppWithCustomOptions(t, false, signer, app.SetupOptions{
			Logger:             log.NewNopLogger(),
			DB:                 dbm.NewMemDB(),
			InvCheckPeriod:     0,
			SkipUpgradeHeights: map[int64]bool{},
			AppOpts:            app.EmptyAppOptions{},
		})
		cdc := bbn.AppCodec()
		err = genutiltest.ExecInitCmd(bbn.BasicModuleManager, home, cdc)
		require.NoError(t, err)

		clientCtx := client.Context{}.
			WithCodec(bbn.AppCodec()).
			WithHomeDir(home).
			WithTxConfig(bbn.TxConfig())

		ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

		fpsToAddFilePath := filepath.Join(home, "fpsToAdd.json")
		fpToAddGen := &btcstktypes.GenesisState{
			FinalityProviders: fpsToAdd,
		}

		jsonBytes, err := cdc.MarshalJSON(fpToAddGen)
		require.NoError(t, err)

		err = tempfile.WriteFileAtomic(fpsToAddFilePath, jsonBytes, 0600)
		require.NoError(t, err)

		cmdSetFp := genhelpers.CmdSetFp()
		cmdSetFp.SetArgs([]string{
			fpsToAddFilePath,
		})
		cmdSetFp.SetContext(ctx)
		err = cmdSetFp.Flags().Set(flags.FlagHome, home)
		require.NoError(t, err)

		// Runs the cmd to write into the genesis
		err = cmdSetFp.Execute()
		require.NoError(t, err)

		cmtcfg, err := genutiltest.CreateDefaultCometConfig(home)
		require.NoError(t, err)

		// Verifies that the new genesis were created
		genFile := cmtcfg.GenesisFile()
		appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
		require.NoError(t, err)

		btcstkGenState := btcstktypes.GenesisStateFromAppState(clientCtx.Codec, appState)
		require.Equal(t, qntFps, len(btcstkGenState.FinalityProviders))

		for i := 0; i < qntFps; i++ {
			bzAdd, err := fpsToAdd[i].Marshal()
			require.NoError(t, err)

			bzGen, err := btcstkGenState.FinalityProviders[i].Marshal()
			require.NoError(t, err)

			require.Equal(t, hex.EncodeToString(bzAdd), hex.EncodeToString(bzGen))
		}

		// tries to add again, it should error out
		err = cmdSetFp.Execute()
		fp := fpsToAdd[0]
		require.EqualError(t, err, fmt.Errorf("error: finality provider: %+v\nwas already set on genesis, or contains the same BtcPk %s than another finality provider", fp, fp.BtcPk.MarshalHex()).Error())
	})
}
