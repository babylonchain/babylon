package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateBTCValidator{}, "btcstaking/MsgCreateBTCValidator", nil)
	cdc.RegisterConcrete(&MsgCreateBTCDelegation{}, "btcstaking/MsgCreateBTCDelegation", nil)
	cdc.RegisterConcrete(&MsgAddCovenantSig{}, "btcstaking/MsgAddCovenantSig", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "btcstaking/MsgUpdateParams", nil)
	cdc.RegisterConcrete(&MsgBTCUndelegate{}, "btcstaking/MsgBtcUndelegate", nil)
	cdc.RegisterConcrete(&MsgAddCovenantUnbondingSigs{}, "btcstaking/MsgAddCovenantUnbondingSigs", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Register messages
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateBTCValidator{},
		&MsgCreateBTCDelegation{},
		&MsgAddCovenantSig{},
		&MsgUpdateParams{},
		&MsgBTCUndelegate{},
		&MsgAddCovenantUnbondingSigs{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
