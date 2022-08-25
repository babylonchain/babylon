package keeper

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetGenBlsKeys registers BLS keys with each validator at genesis
func (k Keeper) SetGenBlsKeys(ctx sdk.Context, genKeys []*types.GenesisKey) {
	for _, key := range genKeys {
		addr, err := sdk.ValAddressFromBech32(key.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		exists := k.RegistrationState(ctx).Exists(addr)
		if exists {
			panic("a validator's BLS key has already been registered")
		}
		ok := key.BlsKey.Pop.IsValid(*key.BlsKey.Pubkey, key.ValPubkey.GetCachedValue().(cryptotypes.PubKey))
		if !ok {
			panic("Proof-of-Possession is not valid")
		}
		err = k.RegistrationState(ctx).CreateRegistration(*key.BlsKey.Pubkey, addr)
		if err != nil {
			panic("failed to register a BLS key")
		}
	}
}
