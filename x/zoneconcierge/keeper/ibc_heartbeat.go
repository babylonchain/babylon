package keeper

import (
	"fmt"
	"time"

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

// SendIBCPacket sends an IBC packet to a channel
// (adapted from https://github.com/cosmos/ibc-go/blob/v5.0.0/modules/apps/transfer/keeper/relay.go)
func (k Keeper) SendIBCPacket(ctx sdk.Context, channel channeltypes.IdentifiedChannel, packetData *types.ZoneconciergePacketData) error {
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
	timeoutTime := uint64(ctx.BlockHeader().Time.Add(time.Hour * 24).UnixNano()) // TODO: parameterise
	zeroheight := clienttypes.ZeroHeight()

	// construct packet from packet data
	packet := channeltypes.NewPacket(
		k.cdc.MustMarshal(packetData),
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		zeroheight,  // no need to set timeout height if timeout timestamp is set
		timeoutTime, // if the packet is not relayed after this time, then the packet will be time out
	)

	// send packet
	if err := k.ics4Wrapper.SendPacket(ctx, channelCap, packet); err != nil {
		// Failed/timeout packet should not make the system crash
		k.Logger(ctx).Error(fmt.Sprintf("failed to send IBC packet (sequence number: %d) to channel %v port %s: %v", packet.Sequence, destinationChannel, destinationPort, err))
	} else {
		k.Logger(ctx).Info(fmt.Sprintf("successfully sent IBC packet (sequence number: %d) to channel %v port %s", packet.Sequence, destinationChannel, destinationPort))
	}

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
