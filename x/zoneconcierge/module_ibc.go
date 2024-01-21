package zoneconcierge

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

type IBCModule struct {
	keeper keeper.Keeper
}

func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	// the IBC channel has to be ordered
	if order != channeltypes.ORDERED {
		return "", errorsmod.Wrapf(channeltypes.ErrInvalidChannelOrdering, "expected %s channel, got %s ", channeltypes.ORDERED, order)
	}

	// Require portID to be the one that ZoneConcierge is bound to
	boundPort := im.keeper.GetPort(ctx)
	if boundPort != portID {
		return "", errorsmod.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}

	// ensure consistency of the protocol version
	if version != types.Version {
		return "", errorsmod.Wrapf(types.ErrInvalidVersion, "got %s, expected %s", version, types.Version)
	}

	// Claim channel capability passed back by IBC module
	if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return "", err
	}

	return version, nil
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	// the IBC channel has to be ordered
	if order != channeltypes.ORDERED {
		return "", errorsmod.Wrapf(channeltypes.ErrInvalidChannelOrdering, "expected %s channel, got %s ", channeltypes.ORDERED, order)
	}

	// Require portID to be the one that ZoneConcierge is bound to
	boundPort := im.keeper.GetPort(ctx)
	if boundPort != portID {
		return "", errorsmod.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}

	// ensure consistency of the protocol version
	if counterpartyVersion != types.Version {
		return "", errorsmod.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: got: %s, expected %s", counterpartyVersion, types.Version)
	}

	// Module may have already claimed capability in OnChanOpenInit in the case of crossing hellos
	// (ie chainA and chainB both call ChanOpenInit before one of them calls ChanOpenTry)
	// If module can already authenticate the capability then module already owns it so we don't need to claim
	// Otherwise, module does not have channel capability and we must claim it from IBC
	if !im.keeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
		// Only claim channel capability passed back by IBC module if we do not already own it
		if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
			return "", err
		}
	}

	return types.Version, nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	_,
	counterpartyVersion string,
) error {
	// check version consistency
	if counterpartyVersion != types.Version {
		return errorsmod.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	}

	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for channels
	return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	// Babylon is supposed to not take any IBC packet
	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	return channeltypes.NewErrorAcknowledgement(errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "Babylon is supposed to not take any IBC packet"))
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	// `x/wasm` uses both protobuf and json to encoded acknowledgement, so we need to try both here
	// - for acknowledgment message with errors defined in `x/wasm`, it uses json
	// - for all other acknowledgement messages, it uses protobuf
	if errProto := types.ModuleCdc.Unmarshal(acknowledgement, &ack); errProto != nil {
		im.keeper.Logger(ctx).Error("cannot unmarshal packet acknowledgement with protobuf", "error", errProto)
		if errJson := types.ModuleCdc.Unmarshal(acknowledgement, &ack); errJson != nil {
			im.keeper.Logger(ctx).Error("cannot unmarshal packet acknowledgement with json", "error", errJson)
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet acknowledgement with protobuf (error: %v) or json (error: %v)", errProto, errJson)
		}
	}

	// // TODO (Babylon): Dispatch and process packet
	// switch packet := modulePacketData.Packet.(type) {
	// default:
	// 	errMsg := fmt.Sprintf("unrecognized %s packet type: %T", types.ModuleName, packet)
	// 	return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	// }

	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		im.keeper.Logger(ctx).Info("received an Acknowledgement message", "result", string(resp.Result))
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeAck,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		im.keeper.Logger(ctx).Error("received an Acknowledgement error message", "error", resp.Error)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeAck,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
			),
		)
	}

	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	var modulePacketData types.ZoneconciergePacketData
	if err := modulePacketData.Unmarshal(modulePacket.GetData()); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}

	// // TODO (Babylon): Dispatch and process packet
	// switch packet := modulePacketData.Packet.(type) {
	// default:
	// 	errMsg := fmt.Sprintf("unrecognized %s packet type: %T", types.ModuleName, packet)
	// 	return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	// }

	// TODO: close channel upon timeout

	return nil
}
