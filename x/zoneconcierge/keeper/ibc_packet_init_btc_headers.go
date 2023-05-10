package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// SendInitBTCHeaders sends w+1 BTC headers to Babylon contract
func (k Keeper) SendInitBTCHeaders(ctx sdk.Context, channel channeltypes.IdentifiedChannel) error {
	// get last w+1 headers
	// GetAscendingTipHeaders will ensure there are enough headers
	w := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout
	headers, err := k.btclcKeeper.GetAscendingTipHeaders(ctx, w+1)
	if err != nil {
		return err
	}

	// wrap BTC timestamp to IBC packet
	packet := &types.ZoneconciergePacketData{
		Packet: &types.ZoneconciergePacketData_InitBtcHeaders{
			InitBtcHeaders: &types.InitBTCHeaders{
				BtcHeaders: headers,
			},
		},
	}

	// send IBC packet
	if err := k.SendIBCPacket(ctx, channel, packet); err != nil {
		k.Logger(ctx).Error("failed to send InitBTCHeaders IBC packet", "channelID", channel.ChannelId, "error", err)
	}

	return nil
}
