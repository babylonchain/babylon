package app

import (
	"cosmossdk.io/log"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	dbm "github.com/cosmos/cosmos-db"

	appparams "github.com/babylonchain/babylon/app/params"
)

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
		EmptyAppOptions{},
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
