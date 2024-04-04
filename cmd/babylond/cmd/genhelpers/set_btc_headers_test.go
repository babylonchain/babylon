package genhelpers_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/babylonchain/babylon/cmd/babylond/cmd/genhelpers"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/testutil/helper"
	btclighttypes "github.com/babylonchain/babylon/x/btclightclient/types"

	"github.com/cosmos/cosmos-sdk/client"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/stretchr/testify/require"
)

func FuzzCmdSetBtcHeaders(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		h := helper.NewHelper(t)
		app, home := h.App, t.TempDir()
		cdc := app.AppCodec()

		err := genutiltest.ExecInitCmd(app.BasicModuleManager, home, cdc)
		require.NoError(t, err)

		cmtcfg, err := genutiltest.CreateDefaultCometConfig(home)
		require.NoError(t, err)

		appState, _, err := genutiltypes.GenesisStateFromGenFile(cmtcfg.GenesisFile())
		require.NoError(t, err)
		btclightGenState := btclighttypes.GenesisStateFromAppState(cdc, appState)

		qntBtcHeaderInGenesis := len(btclightGenState.BtcHeaders)
		require.GreaterOrEqual(t, qntBtcHeaderInGenesis, 1)
		last := btclightGenState.BtcHeaders[qntBtcHeaderInGenesis-1]

		qntBtcHeaders := int(datagen.RandomInt(r, 10)) + 1
		btcHeadersToAdd := make([]*btclighttypes.BTCHeaderInfo, qntBtcHeaders)
		for i := 0; i < qntBtcHeaders; i++ {
			new := datagen.GenRandomBTCHeaderInfoWithParent(r, last)
			btcHeadersToAdd[i] = new
			last = new
		}

		clientCtx := client.Context{}.
			WithCodec(app.AppCodec()).
			WithHomeDir(home).
			WithTxConfig(app.TxConfig())
		ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

		btcHeadersToAddFilePath := filepath.Join(home, "btcHeadersToAdd.json")
		writeFileProto(t, cdc, btcHeadersToAddFilePath, &btclighttypes.GenesisState{
			BtcHeaders: btcHeadersToAdd,
		})

		cmdSetBtcHeaders := genhelpers.CmdSetBtcHeaders()
		cmdSetBtcHeaders.SetArgs([]string{btcHeadersToAddFilePath})
		cmdSetBtcHeaders.SetContext(ctx)

		// Runs the cmd to write into the genesis
		err = cmdSetBtcHeaders.Execute()
		require.NoError(t, err)

		// reloads appstate
		appState, _, err = genutiltypes.GenesisStateFromGenFile(cmtcfg.GenesisFile())
		require.NoError(t, err)
		btclightGenState = btclighttypes.GenesisStateFromAppState(cdc, appState)
		require.Equal(t, qntBtcHeaders+qntBtcHeaderInGenesis, len(btclightGenState.BtcHeaders))

		for i := 0; i < qntBtcHeaders; i++ {
			bzAdd, err := btcHeadersToAdd[i].Marshal()
			require.NoError(t, err)

			bzGen, err := btclightGenState.BtcHeaders[qntBtcHeaderInGenesis+i].Marshal()
			require.NoError(t, err)

			require.Equal(t, hex.EncodeToString(bzAdd), hex.EncodeToString(bzGen))
		}

		// tries to add again, it should error out
		err = cmdSetBtcHeaders.Execute()
		btcHeader := btcHeadersToAdd[0]
		require.EqualError(t, err, fmt.Errorf("error: btc header: %+v\nwas already set on genesis, or contains the same hash %s than another btc header", btcHeader, btcHeader.Hash.MarshalHex()).Error())
	})
}
