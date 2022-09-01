package keeper

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
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
		ok := key.BlsKey.Pop.IsValid(*key.BlsKey.Pubkey, key.ValPubkey)
		if !ok {
			panic("Proof-of-Possession is not valid")
		}
		err = k.RegistrationState(ctx).CreateRegistration(*key.BlsKey.Pubkey, addr)
		if err != nil {
			panic("failed to register a BLS key")
		}
	}
}

// InitGenValLifecycle records genesis validator states into their lifecycles
func (k Keeper) InitGenValLifecycle(ctx sdk.Context, genKeys []*types.GenesisKey) {
	for _, key := range genKeys {
		addr, err := sdk.ValAddressFromBech32(key.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		k.epochingKeeper.InitValState(ctx, addr)                                 // creation request submitted at block 0
		k.epochingKeeper.UpdateValState(ctx, addr, epochingtypes.ValStateBonded) // bonded at block 0
	}
}
