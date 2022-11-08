package zoneconcierge

import (
	"time"

	"github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	for _, channel := range k.GetAllChannels(ctx) {
		k.SendHeartbeatIBCPacket(ctx, channel) // TODO: error handling
	}
}
