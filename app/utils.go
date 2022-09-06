package app

import (
	"github.com/babylonchain/babylon/privval"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	"os"
	"path/filepath"
)

type PrivSigner struct {
	WrappedPV *privval.WrappedFilePV
	ClientCtx client.Context
}

func InitClientContext(clientCtx client.Context, backend string) (*PrivSigner, error) {
	// setup private validator
	nodeCfg := tmconfig.DefaultConfig()
	pvKeyFile := filepath.Join(".testnet/node0/babylond", nodeCfg.PrivValidatorKeyFile())
	err := tmos.EnsureDir(filepath.Dir(pvKeyFile), 0777)
	if err != nil {
		return nil, err
	}
	pvStateFile := filepath.Join(".testnet/node0/babylond", nodeCfg.PrivValidatorStateFile())
	err = tmos.EnsureDir(filepath.Dir(pvStateFile), 0777)
	if err != nil {
		return nil, err
	}
	wrappedPV := privval.LoadOrGenWrappedFilePV(pvKeyFile, pvStateFile)

	// setup client context
	encodingCfg := MakeTestEncodingConfig()
	clientCtx = client.Context{}.
		WithHomeDir(DefaultNodeHome).
		WithInterfaceRegistry(encodingCfg.InterfaceRegistry).
		WithCodec(encodingCfg.Marshaler).
		WithLegacyAmino(encodingCfg.Amino).
		WithTxConfig(encodingCfg.TxConfig).
		WithAccountRetriever(types.AccountRetriever{}).
		WithInput(os.Stdin).
		WithBroadcastMode(flags.BroadcastBlock).
		WithFromAddress(sdk.AccAddress(wrappedPV.GetAddress())).
		WithFeeGranterAddress(sdk.AccAddress(wrappedPV.GetAddress())).
		WithViper("").
		WithFromName("node0").
		WithChainID("chain-test").
		WithSkipConfirmation(true)
	clientCtx, err = config.ReadFromClientConfig(clientCtx)
	if err != nil {
		return nil, err
	}
	clientCtx.KeyringDir = "/Users/lanpo/Worksapce/babylon/.testnet/node0/babylond"
	kb, err := client.NewKeyringFromBackend(clientCtx, backend)

	//kb, err := keyring.New(sdk.KeyringServiceName(), backend, DefaultNodeHome, clientCtx.Input)
	//kb, err := keyring.New(sdk.KeyringServiceName(), backend, keyringPath, clientCtx.Input)
	if err != nil {
		return nil, err
	}
	clientCtx = clientCtx.WithKeyring(kb).WithChainID("chain-test").WithBroadcastMode(flags.BroadcastBlock)

	return &PrivSigner{
		WrappedPV: wrappedPV,
		ClientCtx: clientCtx,
	}, nil
}
