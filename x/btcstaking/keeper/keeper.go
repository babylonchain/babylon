package keeper

import (
	"context"
	"fmt"

	corestoretypes "cosmossdk.io/core/store"

	"cosmossdk.io/log"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/btcstaking/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService corestoretypes.KVStoreService

		btclcKeeper types.BTCLightClientKeeper
		btccKeeper  types.BtcCheckpointKeeper
		ckptKeeper  types.CheckpointingKeeper

		hooks types.BtcStakingHooks

		btcNet *chaincfg.Params
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,

	btclcKeeper types.BTCLightClientKeeper,
	btccKeeper types.BtcCheckpointKeeper,
	ckptKeeper types.CheckpointingKeeper,

	btcNet *chaincfg.Params,
	authority string,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: storeService,

		btclcKeeper: btclcKeeper,
		btccKeeper:  btccKeeper,
		ckptKeeper:  ckptKeeper,

		hooks: nil,

		btcNet:    btcNet,
		authority: authority,
	}
}

// SetHooks sets the BTC staking hooks
func (k *Keeper) SetHooks(sh types.BtcStakingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set BTC staking hooks twice")
	}

	k.hooks = sh

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// BeginBlocker is invoked upon `BeginBlock` of the system. The function
// iterates over all BTC delegations under non-slashed finality providers
// to 1) record the voting power table for the current height, and 2) record
// the voting power distribution cache used for computing voting power table
// and distributing rewards once the block is finalised by finality providers.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	// index BTC height at the current height
	k.IndexBTCHeight(ctx)
	// update voting power distribution
	k.UpdatePowerDist(ctx)

	return nil
}

func (k Keeper) GetLastFinalizedEpoch(ctx context.Context) uint64 {
	return k.ckptKeeper.GetLastFinalizedEpoch(ctx)
}
