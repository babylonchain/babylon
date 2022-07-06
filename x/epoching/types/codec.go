package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgWrappedDelegate{}, "epoching/WrappedDelegate", nil)
	cdc.RegisterConcrete(&MsgWrappedUndelegate{}, "epoching/WrappedUndelegate", nil)
	cdc.RegisterConcrete(&MsgWrappedBeginRedelegate{}, "epoching/WrappedBeginRedelegate", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Register messages
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgWrappedDelegate{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgWrappedUndelegate{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgWrappedBeginRedelegate{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
