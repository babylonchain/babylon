package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddBlsSig{}, "checkpointing/AddBlsSig", nil)
	cdc.RegisterConcrete(&MsgWrappedCreateValidator{}, "checkpointing/WrappedCreateValidator", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {

	// Register messages
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddBlsSig{},
		&MsgWrappedCreateValidator{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
