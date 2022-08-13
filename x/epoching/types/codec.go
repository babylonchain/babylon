package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgWrappedDelegate{}, "epoching/WrappedDelegate", nil)
	cdc.RegisterConcrete(&MsgWrappedUndelegate{}, "epoching/WrappedUndelegate", nil)
	cdc.RegisterConcrete(&MsgWrappedBeginRedelegate{}, "epoching/WrappedBeginRedelegate", nil)
	cdc.RegisterConcrete(&QueuedMessage{}, "epoching/QueuedMessage", nil)
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
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&QueuedMessage{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/staking module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/staking and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterLegacyAminoCodec(Amino)
	cryptocodec.RegisterCrypto(Amino)
	Amino.Seal()
}
