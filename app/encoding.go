package app

import (
	"os"

	"cosmossdk.io/log"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	simsutils "github.com/cosmos/cosmos-sdk/testutil/sims"

	appparams "github.com/babylonchain/babylon/app/params"
	bbn "github.com/babylonchain/babylon/types"
)

// TmpAppOptions returns an app option with tmp dir and btc network
func TmpAppOptions() simsutils.AppOptionsMap {
	dir, err := os.MkdirTemp("", "babylon-tmp-app")
	if err != nil {
		panic(err)
	}
	appOpts := simsutils.AppOptionsMap{
		flags.FlagHome:       dir,
		"btc-config.network": string(bbn.BtcSimnet),
	}
	return appOpts
}

func NewTmpBabylonApp() *BabylonApp {
	signer, _ := SetupTestPrivSigner()
	return NewBabylonApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		0,
		signer,
		TmpAppOptions(),
		[]wasmkeeper.Option{})
}

// GetEncodingConfig returns a *registered* encoding config
// Note that the only way to register configuration is through the app creation
func GetEncodingConfig() *appparams.EncodingConfig {
	tmpApp := NewTmpBabylonApp()
	return &appparams.EncodingConfig{
		InterfaceRegistry: tmpApp.InterfaceRegistry(),
		Codec:             tmpApp.AppCodec(),
		TxConfig:          tmpApp.TxConfig(),
		Amino:             tmpApp.LegacyAmino(),
	}
}
