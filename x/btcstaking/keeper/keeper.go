package keeper

import (
	"context"
	"fmt"

	corestoretypes "cosmossdk.io/core/store"

	"cosmossdk.io/log"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService corestoretypes.KVStoreService

		btclcKeeper types.BTCLightClientKeeper
		btccKeeper  types.BtcCheckpointKeeper

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

	btcNet *chaincfg.Params,
	authority string,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: storeService,

		btclcKeeper: btclcKeeper,
		btccKeeper:  btccKeeper,

		btcNet:    btcNet,
		authority: authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) BeginBlocker(ctx context.Context) error {
	// index BTC height at the current height
	k.IndexBTCHeight(ctx)

	covenantQuorum := k.GetParams(ctx).CovenantQuorum
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// prepare for recording finality providers with positive voting power
	// key is the finality provider's FP BTC PK hex, and value is the
	// voting power
	fpPowerMap := map[string]uint64{}
	// prepare for recording finality providers and their BTC delegations
	// for rewards
	fpDistMap := map[string]*types.FinalityProviderDistInfo{}

	k.IterateActiveFPsAndBTCDelegations(
		ctx,
		func(fp *types.FinalityProvider, btcDel *types.BTCDelegation) {
			fpBTCPKHex := fp.BtcPk.MarshalHex()

			// record active finality providers
			power := btcDel.VotingPower(btcTipHeight, wValue, covenantQuorum)
			if power == 0 {
				return // skip if no voting power
			}
			fpPowerMap[fpBTCPKHex] += power

			// create fp dist info if not exist
			if _, ok := fpDistMap[fpBTCPKHex]; !ok {
				fpDistMap[fpBTCPKHex] = types.NewFinalityProviderDistInfo(fp)
			}
			// append BTC delegation
			fpDistMap[fpBTCPKHex].AddBTCDel(btcDel, btcTipHeight, wValue, covenantQuorum)
		},
	)

	// return directly if there is no active finality provider
	if len(fpPowerMap) == 0 {
		return nil
	}

	// get top N finality providers and set their voting power to KV store
	k.setCurrentTopNVotingPower(ctx, fpPowerMap)

	// create reward distribution cache
	rdc := types.NewRewardDistCache()
	for _, fpDistInfo := range fpDistMap {
		// try to add this finality provider distribution info to reward distribution cache
		rdc.AddFinalityProviderDistInfo(fpDistInfo)
	}

	// all good, set the reward distribution cache of the current height
	k.setRewardDistCache(ctx, uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height), rdc)

	return nil
}
