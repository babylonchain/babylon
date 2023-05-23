package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

func (k Keeper) GetAllChannels(ctx sdk.Context) []channeltypes.IdentifiedChannel {
	return k.channelKeeper.GetAllChannels(ctx)
}

// GetAllOpenZCChannels returns all open channels that are connected to ZoneConcierge's port
func (k Keeper) GetAllOpenZCChannels(ctx sdk.Context) []channeltypes.IdentifiedChannel {
	zcPort := k.GetPort(ctx)
	channels := k.GetAllChannels(ctx)

	openZCChannels := []channeltypes.IdentifiedChannel{}
	for _, channel := range channels {
		if channel.State != channeltypes.OPEN {
			continue
		}
		if channel.PortId != zcPort {
			continue
		}
		openZCChannels = append(openZCChannels, channel)
	}

	return openZCChannels
}

// isChannelUninitialized checks whether the channel is not initilialised yet
// it's done by checking whether the packet sequence number is 1 (the first sequence number) or not
func (k Keeper) isChannelUninitialized(ctx sdk.Context, channel channeltypes.IdentifiedChannel) bool {
	portID := channel.PortId
	channelID := channel.ChannelId
	// NOTE: channeltypes.IdentifiedChannel object is guaranteed to exist, so guaranteed to be found
	nextSeqSend, _ := k.channelKeeper.GetNextSequenceSend(ctx, portID, channelID)
	return nextSeqSend == 1
}
