package app

import (
	"github.com/cosmos/cosmos-sdk/std"

	appparams "github.com/babylonchain/babylon/app/params"
)

var encodingConfig appparams.EncodingConfig = MakeEncodingConfig()

func GetEncodingConfig() appparams.EncodingConfig {
	return encodingConfig
}

// MakeEncodingConfig creates an EncodingConfig.
func MakeEncodingConfig() appparams.EncodingConfig {
	encodingConfig := appparams.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
