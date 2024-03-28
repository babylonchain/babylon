package genhelpers_test

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	cmtconfig "github.com/cometbft/cometbft/config"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/cmd/babylond/cmd/genhelpers"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

func Test_CmdCreateBls(t *testing.T) {
	home := t.TempDir()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	signer, err := app.SetupTestPrivSigner()
	require.NoError(t, err)
	bbn := app.NewBabylonAppWithCustomOptions(t, false, signer, app.SetupOptions{
		Logger:             logger,
		DB:                 dbm.NewMemDB(),
		InvCheckPeriod:     0,
		SkipUpgradeHeights: map[int64]bool{},
		AppOpts:            app.EmptyAppOptions{},
	})
	err = genutiltest.ExecInitCmd(bbn.BasicModuleManager, home, bbn.AppCodec())
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.
		WithCodec(bbn.AppCodec()).
		WithHomeDir(home).
		WithTxConfig(bbn.TxConfig())

	ctx := context.Background()
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	genBlsCmd := genhelpers.CmdCreateBls()
	genBlsCmd.SetArgs([]string{fmt.Sprintf("--%s=%s", flags.FlagHome, home)})

	// create keyring to get the validator address
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, home, bufio.NewReader(genBlsCmd.InOrStdin()), clientCtx.Codec)
	require.NoError(t, err)
	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	require.NoError(t, err)
	addr, _, err := testutil.GenerateSaveCoinKey(kb, home, "", true, algo)
	require.NoError(t, err)

	// create BLS keys
	nodeCfg := cmtconfig.DefaultConfig()
	keyPath := filepath.Join(home, nodeCfg.PrivValidatorKeyFile())
	statePath := filepath.Join(home, nodeCfg.PrivValidatorStateFile())
	filePV := privval.GenWrappedFilePV(keyPath, statePath)
	defer filePV.Clean(keyPath, statePath)
	filePV.SetAccAddress(addr)

	// execute the gen-bls cmd
	err = genBlsCmd.ExecuteContext(ctx)
	require.NoError(t, err)
	outputFilePath := filepath.Join(filepath.Dir(keyPath), fmt.Sprintf("gen-bls-%s.json", sdk.ValAddress(addr).String()))
	require.NoError(t, err)
	genKey, err := types.LoadGenesisKeyFromFile(outputFilePath)
	require.NoError(t, err)
	require.Equal(t, sdk.ValAddress(addr).String(), genKey.ValidatorAddress)
	require.True(t, filePV.Key.BlsPubKey.Equal(*genKey.BlsKey.Pubkey))
	require.Equal(t, filePV.Key.PubKey.Bytes(), genKey.ValPubkey.Bytes())
	require.True(t, genKey.BlsKey.Pop.IsValid(*genKey.BlsKey.Pubkey, genKey.ValPubkey))
}
