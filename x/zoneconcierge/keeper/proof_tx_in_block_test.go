package keeper_test

import (
	"fmt"
	"testing"

	"github.com/babylonchain/babylon/app"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestProveTxInBlock(t *testing.T) {
	// setup virtual network
	cfg := network.DefaultConfig()
	encodingCfg := app.MakeTestEncodingConfig()
	cfg.InterfaceRegistry = encodingCfg.InterfaceRegistry
	cfg.TxConfig = encodingCfg.TxConfig
	cfg.NumValidators = 2
	cfg.RPCAddress = "tcp://0.0.0.0:26657" // TODO: parameterise this
	testNetwork, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	defer testNetwork.Cleanup()

	// enter block 1 so that gentxs take effect and validators have tokens
	err = testNetwork.WaitForNextBlock()
	require.NoError(t, err)

	_, babylonChain, _, zcKeeper := SetupTest(t)
	ctx := babylonChain.GetContext()

	// construct client context
	val := testNetwork.Validators[0]
	val.ClientCtx.FromAddress = val.Address
	val.ClientCtx.FeePayer = val.Address
	val.ClientCtx.FeeGranter = val.Address
	require.NotEmpty(t, val.Address, val.ValAddress)
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.String(flags.FlagFees, "", "Fees to pay along with transaction; eg: 10ubbn")
	fee := fmt.Sprintf("100%s", sdk.DefaultBondDenom)
	err = fs.Set(flags.FlagFees, fee)
	require.NoError(t, err)

	// construct a simple msg
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10)))
	msg := banktypes.NewMsgSend(val.Address, testNetwork.Validators[1].Address, coins)

	// construct a tx that includes this msg
	txf := tx.NewFactoryCLI(val.ClientCtx, fs).
		WithTxConfig(val.ClientCtx.TxConfig).
		WithAccountRetriever(val.ClientCtx.AccountRetriever)
	txf, err = txf.Prepare(val.ClientCtx)
	require.NoError(t, err)
	txb, err := txf.BuildUnsignedTx(msg)
	require.NoError(t, err)
	keys, err := val.ClientCtx.Keyring.List()
	require.NoError(t, err)
	err = tx.Sign(txf.WithKeybase(val.ClientCtx.Keyring), keys[0].Name, txb, true)
	require.NoError(t, err)
	txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txb.GetTx())
	require.NoError(t, err)

	// submit the tx to Tendermint
	resp, err := val.RPCClient.BroadcastTxSync(ctx, txBytes)
	require.NoError(t, err)
	txHash := resp.Hash

	err = testNetwork.WaitForNextBlock()
	require.NoError(t, err)
	err = testNetwork.WaitForNextBlock()
	require.NoError(t, err)

	proof, err := zcKeeper.ProveTxInBlock(ctx, txHash)
	require.NoError(t, err)

	err = zckeeper.VerifyTxInBlock(txHash, proof)
	require.NoError(t, err)
}
