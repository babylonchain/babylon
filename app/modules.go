package app

import (
	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/client"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
)

var _ ibctesting.ChainApp = &BabylonApp{}

// The following functions are required by ibctesting
// (copied from https://github.com/osmosis-labs/osmosis/blob/main/app/modules.go)

func (app *BabylonApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

func (app *BabylonApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

func (app *BabylonApp) GetBankKeeper() bankkeeper.Keeper {
	return app.BankKeeper
}

func (app *BabylonApp) GetStakingKeeper() *stakingkeeper.Keeper {
	return app.StakingKeeper
}

func (app *BabylonApp) GetAccountKeeper() authkeeper.AccountKeeper {
	return app.AccountKeeper
}

func (app *BabylonApp) GetWasmKeeper() wasmkeeper.Keeper {
	return app.WasmKeeper
}

func (app *BabylonApp) GetTxConfig() client.TxConfig {
	return app.TxConfig()
}
