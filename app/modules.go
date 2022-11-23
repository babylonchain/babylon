package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v5/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v5/testing/types"
)

// The following functions are required by ibctesting
// (copied from https://github.com/osmosis-labs/osmosis/blob/main/app/modules.go)

func (app *BabylonApp) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

func (app *BabylonApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper // This is a *ibckeeper.Keeper
}

func (app *BabylonApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

func (app *BabylonApp) GetTxConfig() client.TxConfig {
	return MakeTestEncodingConfig().TxConfig
}
