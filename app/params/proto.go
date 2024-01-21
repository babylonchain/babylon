//go:build !test_amino
// +build !test_amino

package params

import (
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// DefaultEncodingConfig returns the default encoding config
func DefaultEncodingConfig() *EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: Bech32PrefixAccAddr,
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: Bech32PrefixValAddr,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	return &EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}
