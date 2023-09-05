package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		bankKeeper     types.BankKeeper
		accountKeeper  types.AccountKeeper
		epochingKeeper types.EpochingKeeper
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
		// name of the FeeCollector ModuleAccount
		feeCollectorName string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	epochingKeeper types.EpochingKeeper,
	authority string,
	feeCollectorName string,
) Keeper {
	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		memKey:           memKey,
		bankKeeper:       bankKeeper,
		accountKeeper:    accountKeeper,
		epochingKeeper:   epochingKeeper,
		authority:        authority,
		feeCollectorName: feeCollectorName,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
