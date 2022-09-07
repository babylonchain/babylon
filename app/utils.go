package app

import (
	"github.com/babylonchain/babylon/privval"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	"path/filepath"
)

type PrivSigner struct {
	WrappedPV *privval.WrappedFilePV
	ClientCtx client.Context
}

func InitPrivSigner(clientCtx client.Context, nodeDir string, backEnd string) (*PrivSigner, error) {
	// setup private validator
	nodeCfg := tmconfig.DefaultConfig()
	pvKeyFile := filepath.Join(nodeDir, nodeCfg.PrivValidatorKeyFile())
	err := tmos.EnsureDir(filepath.Dir(pvKeyFile), 0777)
	if err != nil {
		return nil, err
	}
	pvStateFile := filepath.Join(nodeDir, nodeCfg.PrivValidatorStateFile())
	err = tmos.EnsureDir(filepath.Dir(pvStateFile), 0777)
	if err != nil {
		return nil, err
	}
	wrappedPV := privval.LoadOrGenWrappedFilePV(pvKeyFile, pvStateFile)

	// setup client context
	encodingCfg := MakeTestEncodingConfig()
	kb, err := client.NewKeyringFromBackend(clientCtx, backEnd)
	if err != nil {
		return nil, err
	}
	clientCtx = clientCtx.
		WithInterfaceRegistry(encodingCfg.InterfaceRegistry).
		WithCodec(encodingCfg.Marshaler).
		WithLegacyAmino(encodingCfg.Amino).
		WithTxConfig(encodingCfg.TxConfig).
		WithAccountRetriever(types.AccountRetriever{}).
		WithFromAddress(sdk.AccAddress(wrappedPV.GetAddress())).
		WithFeeGranterAddress(sdk.AccAddress(wrappedPV.GetAddress())).
		WithSkipConfirmation(true).
		WithKeyring(kb)

	return &PrivSigner{
		WrappedPV: wrappedPV,
		ClientCtx: clientCtx,
	}, nil
}
