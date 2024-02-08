package keeper

import (
	corestoretypes "cosmossdk.io/core/store"
	"fmt"

	"cosmossdk.io/log"
	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService corestoretypes.KVStoreService

		epochingKeeper types.EpochingKeeper
		bankKeeper     types.BankKeeper
		accountKeeper  types.AccountKeeper
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
		// name of the FeeCollector ModuleAccount
		feeCollectorName string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	epochingKeeper types.EpochingKeeper,
	authority string,
	feeCollectorName string,
) Keeper {
	return Keeper{
		cdc:              cdc,
		storeService:     storeService,
		epochingKeeper:   epochingKeeper,
		bankKeeper:       bankKeeper,
		accountKeeper:    accountKeeper,
		authority:        authority,
		feeCollectorName: feeCollectorName,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
