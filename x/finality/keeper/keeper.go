package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/finality/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService corestoretypes.KVStoreService

		BTCStakingKeeper types.BTCStakingKeeper
		IncentiveKeeper  types.IncentiveKeeper
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		hooks types.FinalityHooks

		// FinalityProviderSigningTracker key: BIP340PubKey bytes | value: FinalityProviderSigningInfo
		FinalityProviderSigningTracker collections.Map[[]byte, types.FinalityProviderSigningInfo]
		// FinalityProviderMissedBlockBitmap key: BIP340PubKey bytes | value: byte key for a finality provider's missed block bitmap chunk
		FinalityProviderMissedBlockBitmap collections.Map[collections.Pair[[]byte, uint64], []byte]
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,
	btcstakingKeeper types.BTCStakingKeeper,
	incentiveKeeper types.IncentiveKeeper,
	authority string,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return Keeper{
		cdc:          cdc,
		storeService: storeService,

		BTCStakingKeeper: btcstakingKeeper,
		IncentiveKeeper:  incentiveKeeper,
		authority:        authority,
		FinalityProviderSigningTracker: collections.NewMap(
			sb,
			types.FinalityProviderSigningInfoKeyPrefix,
			"finality_provider_signing_info",
			collections.BytesKey,
			codec.CollValue[types.FinalityProviderSigningInfo](cdc),
		),
		FinalityProviderMissedBlockBitmap: collections.NewMap(
			sb,
			types.FinalityProviderMissedBlockBitmapKeyPrefix,
			"finality_provider_missed_block_bitmap",
			collections.PairKeyCodec(collections.BytesKey, collections.Uint64Key),
			collections.BytesValue,
		),
	}
}

// SetHooks sets the finality hooks
func (k *Keeper) SetHooks(sh types.FinalityHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set finality hooks twice")
	}

	k.hooks = sh

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetLastFinalizedEpoch(ctx context.Context) uint64 {
	return k.BTCStakingKeeper.GetLastFinalizedEpoch(ctx)
}
