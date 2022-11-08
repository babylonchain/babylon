package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
)

// SendHeartbeatIBCPacket sends an empty IBC packet to a channel
// Doing this periodically keeps the relayer awake to relay headers
// (adapted from https://github.com/cosmos/ibc-go/blob/v5.0.0/modules/apps/transfer/keeper/relay.go)
func (k Keeper) SendHeartbeatIBCPacket(ctx sdk.Context, channel channeltypes.IdentifiedChannel) {
	// sourcePort := channel.PortId
	// sourceChannel := channel.ChannelId

	// destinationPort := channel.Counterparty.GetPortID()
	// destinationChannel := channel.Counterparty.GetChannelID()

	// sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	// if !found {
	// 	return sdkerrors.Wrapf(
	// 		channeltypes.ErrSequenceSendNotFound,
	// 		"source port: %s, source channel: %s", sourcePort, sourceChannel,
	// 	)
	// }

	// // begin createOutgoingPacket logic
	// // See spec for this logic: https://github.com/cosmos/ibc/tree/master/spec/app/ics-020-fungible-token-transfer#packet-relay
	// channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	// if !ok {
	// 	return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	// }
}
