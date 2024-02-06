package keeper

import (
	"context"
	"fmt"
	"sort"

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

// BeginBlocker is invoked upon `BeginBlock` of the system. The function
// iterates over all BTC delegations under non-slashed finality providers
// to 1) record the voting power table for the current height, and 2) record
// the reward distribution cache used for distributing rewards once the block
// is finalised by finality providers.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	// index BTC height at the current height
	k.IndexBTCHeight(ctx)

	covenantQuorum := k.GetParams(ctx).CovenantQuorum
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// prepare for recording finality providers and their BTC delegations
	// for rewards
	fpDistMap := map[string]*types.FinalityProviderDistInfo{}

	k.IterateActiveFPsAndBTCDelegations(
		ctx,
		func(fp *types.FinalityProvider, btcDel *types.BTCDelegation) bool {
			fpBTCPKHex := fp.BtcPk.MarshalHex()

			// create fp dist info if not exist
			if _, ok := fpDistMap[fpBTCPKHex]; !ok {
				fpDistMap[fpBTCPKHex] = types.NewFinalityProviderDistInfo(fp)
			}
			// append BTC delegation
			fpDistMap[fpBTCPKHex].AddBTCDel(btcDel, btcTipHeight, wValue, covenantQuorum)
			return true
		},
	)

	var distInfoSlice []*types.FinalityProviderDistInfo = make([]*types.FinalityProviderDistInfo, len(fpDistMap))
	var i = 0
	for _, distMap := range fpDistMap {
		distInfoSlice[i] = distMap
		i++
	}

	// sort the slice
	sort.SliceStable(distInfoSlice, func(i, j int) bool {
		return distInfoSlice[i].TotalVotingPower > distInfoSlice[j].TotalVotingPower
	})

	maxProviders := int(k.GetParams(ctx).MaxActiveFinalityProviders)
	currentNumberOfActiveProviders := len(distInfoSlice)

	var toCheck int
	if currentNumberOfActiveProviders >= maxProviders {
		toCheck = maxProviders
	} else {
		toCheck = currentNumberOfActiveProviders
	}
	// get current Babylon height
	babylonTipHeight := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	rdc := types.NewRewardDistCache()
	for i := 0; i < toCheck; i++ {
		distInfo := distInfoSlice[i]
		k.SetVotingPower(ctx, distInfo.BtcPk.MustMarshal(), babylonTipHeight, distInfo.TotalVotingPower)
		rdc.AddFinalityProviderDistInfo(distInfo)
	}

	// all good, set the reward distribution cache of the current height
	k.setRewardDistCache(ctx, uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height), rdc)

	return nil
}
