package cmd_test

import (
	"bufio"
	"context"
	"fmt"
	"github.com/babylonchain/babylon/cmd/babylond/cmd"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmconfig "github.com/tendermint/tendermint/config"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/babylonchain/babylon/app"
)

// pre-conditions of running gen-bls cmd are
// 1. creation of keyring to get the validator address
// 2. creation of BLS keys (bls_pub_key and pop)
func Test_GenBlsCmd(t *testing.T) {
	home := t.TempDir()
	encodingConfig := app.MakeTestEncodingConfig()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
	require.NoError(t, err)

	err = genutiltest.ExecInitCmd(app.ModuleBasics, home, encodingConfig.Marshaler)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithHomeDir(home).
		WithTxConfig(encodingConfig.TxConfig)

	ctx := context.Background()
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	genBlsCmd := cmd.GenBlsCmd()
	genBlsCmd.SetArgs([]string{fmt.Sprintf("--%s=%s", flags.FlagHome, home)})

	// create keyring to get the validator address
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, home, bufio.NewReader(genBlsCmd.InOrStdin()))
	require.NoError(t, err)
	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	require.NoError(t, err)
	addr, _, err := testutil.GenerateSaveCoinKey(kb, home, "", true, algo)
	require.NoError(t, err)

	// create BLS keys
	nodeCfg := tmconfig.DefaultConfig()
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
