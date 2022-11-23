package zoneconcierge

import (
	"time"

	"github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker sends a pending packet for every channel upon each new block,
// so that the relayer is kept awake to relay headers
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	for _, channel := range k.GetAllChannels(ctx) {
		if channel.State == channeltypes.OPEN {
			if err := k.SendHeartbeatIBCPacket(ctx, channel); err != nil {
				panic(err)
			}
		}
	}

	return []abci.ValidatorUpdate{}
}
