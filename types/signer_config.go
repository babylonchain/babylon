package types

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
)

func ParseKeyNameFromConfig(opts servertypes.AppOptions) string {
	valueInterface := opts.Get("signer-config.key-name")
	if valueInterface == nil {
		panic("Signer key name should be provided in options")
	}
	keyName, err := cast.ToStringE(valueInterface)
	if err != nil {
		panic("Signer key name should be valid string")
	}

	return keyName
}
