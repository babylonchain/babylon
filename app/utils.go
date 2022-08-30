package app

import (
	"github.com/babylonchain/babylon/privval"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	tmconfig "github.com/tendermint/tendermint/config"
	"os"
)

type PrivSigner struct {
	WrappedPV *privval.WrappedFilePV
	ClientCtx client.Context
}

func InitClientContext(clientCtx client.Context, backend string) (*PrivSigner, error) {
	nodeCfg := tmconfig.DefaultConfig()
	wrappedPV := privval.LoadOrGenWrappedFilePV(nodeCfg.PrivValidatorKeyFile(), nodeCfg.PrivValidatorStateFile())
	encodingCfg := MakeTestEncodingConfig()
	clientCtx = client.Context{}.
		WithHomeDir(DefaultNodeHome).
		WithInterfaceRegistry(encodingCfg.InterfaceRegistry).
		WithCodec(encodingCfg.Marshaler).
		WithLegacyAmino(encodingCfg.Amino).
		WithTxConfig(encodingCfg.TxConfig).
		WithAccountRetriever(types.AccountRetriever{}).
		WithInput(os.Stdin).
		WithBroadcastMode(flags.BroadcastAsync).
		WithFromAddress(sdk.AccAddress(wrappedPV.GetAddress()))

	kb, err := keyring.New(sdk.KeyringServiceName(), backend, DefaultNodeHome, clientCtx.Input)
	if err != nil {
		return nil, err
	}
	clientCtx.WithKeyring(kb)

	return &PrivSigner{
		WrappedPV: wrappedPV,
		ClientCtx: clientCtx,
	}, nil
}
