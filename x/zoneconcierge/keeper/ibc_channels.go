package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
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

func (k Keeper) AddUninitedChannel(ctx sdk.Context, channelID string) {
	store := k.ibcChannelsStore(ctx)
	store.Set([]byte(channelID), []byte{0x00})
}

func (k Keeper) afterChannelInited(ctx sdk.Context, channelID string) {
	store := k.ibcChannelsStore(ctx)
	store.Delete([]byte(channelID))
}

func (k Keeper) isChannelUninited(ctx sdk.Context, channelID string) bool {
	store := k.ibcChannelsStore(ctx)
	return store.Has([]byte(channelID))
}

// ibcChannelsStore stores initialisation status of IBC channels
// prefix: EpochChainInfoKey
// key: channel ID
// value: nil
func (k Keeper) ibcChannelsStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.IBCChannelsKey)
}
