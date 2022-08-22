package keeper

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetGenBlsKeys registers BLS keys with each validator at genesis
func (k Keeper) SetGenBlsKeys(ctx sdk.Context, blsKeys []*types.BlsKey) {
	for _, blskey := range blsKeys {
		valAddr, err := sdk.ValAddressFromHex(blskey.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		exists := k.RegistrationState(ctx).Exists(valAddr)
		if exists {
			panic("a validator's BLS key has already been registered")
		}
		ok := blskey.Pop.IsValid(*blskey.Pubkey)
		if !ok {
			panic("Proof-of-Possession is not valid")
		}
		err = k.RegistrationState(ctx).CreateRegistration(*blskey.Pubkey, valAddr)
		if err != nil {
			panic("failed to register a BLS key")
		}
	}
}
