package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService corestoretypes.KVStoreService
		hooks        types.EpochingHooks
		bk           types.BankKeeper
		stk          types.StakingKeeper
		router       *baseapp.MsgServiceRouter
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,
	bk types.BankKeeper,
	stk types.StakingKeeper,
	authority string,
) Keeper {

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		hooks:        nil,
		bk:           bk,
		stk:          stk,
		authority:    authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetHooks sets the validator hooks
func (k *Keeper) SetHooks(eh types.EpochingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}

	k.hooks = eh

	return k
}

// SetMsgServiceRouter sets the msgServiceRouter
func (k *Keeper) SetMsgServiceRouter(router *baseapp.MsgServiceRouter) *Keeper {
	k.router = router
	return k
}
