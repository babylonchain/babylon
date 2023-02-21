package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func GetSignBytes(epoch uint64, hash []byte) []byte {
	return append(sdk.Uint64ToBigEndian(epoch), hash...)
}
