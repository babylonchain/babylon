package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc            codec.BinaryCodec
		storeKey       sdk.StoreKey
		memKey         sdk.StoreKey
		stakingKeeper  types.StakingKeeper
		epochingKeeper types.EpochingKeeper
		paramstore     paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	stakingKeeper types.StakingKeeper,
	epochingKeeper types.EpochingKeeper,
	ps paramtypes.Subspace,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		memKey:         memKey,
		stakingKeeper:  stakingKeeper,
		epochingKeeper: epochingKeeper,
		paramstore:     ps,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) AddBlsSig(ctx sdk.Context, sig *types.BlsSig) error {
	// TODO: some checks: 1. duplication check 2. epoch check 3. raw ckpt existence check
	// TODO: aggregate bls sigs and try to build raw checkpoints
	k.BlsSigsState(ctx).InsertBlsSig(sig)
	return nil
}

func (k Keeper) AddCheckpoint(ctx sdk.Context, epoch uint64, ckpt *types.RawCheckpoint) error {
	panic("implement this")
}
