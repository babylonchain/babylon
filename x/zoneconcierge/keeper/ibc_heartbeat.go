package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	metrics "github.com/armon/go-metrics"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"
	coretypes "github.com/cosmos/ibc-go/v5/modules/core/types"
)

// SendHeartbeatIBCPacket sends an empty IBC packet to a channel
// Doing this periodically keeps the relayer awake to relay headers
// (adapted from https://github.com/cosmos/ibc-go/blob/v5.0.0/modules/apps/transfer/keeper/relay.go)
func (k Keeper) SendHeartbeatIBCPacket(ctx sdk.Context, channel channeltypes.IdentifiedChannel) error {
	// get src/dst ports and channels
	sourcePort := channel.PortId
	sourceChannel := channel.ChannelId
	destinationPort := channel.Counterparty.GetPortID()
	destinationChannel := channel.Counterparty.GetChannelID()

	// find the next sequence number
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", sourcePort, sourceChannel,
		)
	}

	// begin createOutgoingPacket logic
	// See spec for this logic: https://github.com/cosmos/ibc/tree/master/spec/app/ics-020-fungible-token-transfer#packet-relay
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrapf(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability: sourcePort: %s, sourceChannel: %s", sourcePort, sourceChannel)
	}

	// timeout
	curHeight := clienttypes.GetSelfHeight(ctx)
	curHeight.RevisionHeight += 100 // TODO: parameterise timeout

	// construct packet
	// note that the data is not allowed to be empty
	packetData := &types.Heartbeat{Msg: "hello"} // TODO: what to send for heartbeat packet?
	packet := channeltypes.NewPacket(
		k.cdc.MustMarshal(packetData),
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		curHeight, // if the packet is not relayed after thit height, then the packet will be timeout
		uint64(0), // no need to set timeout timestamp if timeout height is set
	)

	// send packet
	if err := k.ics4Wrapper.SendPacket(ctx, channelCap, packet); err != nil {
		return err
	}
	k.Logger(ctx).Info("successfully sent heartbeat IBC packet to channel %v port %s", destinationChannel, destinationPort)

	// metrics stuff
	labels := []metrics.Label{
		telemetry.NewLabel(coretypes.LabelDestinationPort, destinationPort),
		telemetry.NewLabel(coretypes.LabelDestinationChannel, destinationChannel),
	}
	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"ibc", types.ModuleName, "send"},
			1,
			labels,
		)
	}()

	return nil
}
