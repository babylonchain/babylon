package keeper

import (
	"context"
	"fmt"
	"sort"

	corestoretypes "cosmossdk.io/core/store"

	"cosmossdk.io/log"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
func (k Keeper) BeginBlockerOld(ctx context.Context) error {
	// index BTC height at the current height
	k.IndexBTCHeight(ctx)

	covenantQuorum := k.GetParams(ctx).CovenantQuorum
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// prepare for recording finality providers with positive voting power
	activeFps := []*types.FinalityProviderWithMeta{}
	// prepare for recording finality providers and their BTC delegations
	// for rewards
	rdc := types.NewRewardDistCache()

	// iterate over all finality providers to find out non-slashed ones that have
	// positive voting power
	k.IterateActiveFPs(
		ctx,
		func(fp *types.FinalityProvider) bool {
			fpDistInfo := types.NewFinalityProviderDistInfo(fp)

			// iterate over all BTC delegations under the finality provider
			// in order to accumulate voting power and reward dist info for it
			k.IterateBTCDelegations(ctx, fp.BtcPk, func(btcDel *types.BTCDelegation) bool {
				// accumulate voting power and reward distribution cache
				fpDistInfo.AddBTCDel(btcDel, btcTipHeight, wValue, covenantQuorum)
				return true
			})

			if fpDistInfo.TotalVotingPower > 0 {
				activeFP := &types.FinalityProviderWithMeta{
					BtcPk:       fp.BtcPk,
					VotingPower: fpDistInfo.TotalVotingPower,
				}
				activeFps = append(activeFps, activeFP)
				rdc.AddFinalityProviderDistInfo(fpDistInfo)
			}

			return true
		},
	)

	// filter out top `MaxActiveFinalityProviders` active finality providers in terms of voting power
	activeFps = types.FilterTopNFinalityProviders(activeFps, k.GetParams(ctx).MaxActiveFinalityProviders)
	// set voting power table
	babylonTipHeight := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	for _, fp := range activeFps {
		k.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), babylonTipHeight, fp.VotingPower)
	}

	// set the reward distribution cache of the current height
	// TODO: only give rewards to top N finality providers and their BTC delegations
	k.setRewardDistCache(ctx, uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height), rdc)

	return nil
}

type FpInfo struct {
	distInfo *types.FinalityProviderDistInfo
	slashed  bool
}

// new not consensus compatible BeginBlocker
func (k Keeper) BeginBlockerNew(ctx context.Context) error {
	k.IndexBTCHeight(ctx)

	covenantQuorum := k.GetParams(ctx).CovenantQuorum
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	finalityProviders := make(map[string]*FpInfo)

	k.IterateBTCDels(ctx, func(delegation *types.BTCDelegation) bool {
		finalityProviderKey := delegation.FpBtcPkList[0].MarshalHex()

		fpData, found := finalityProviders[finalityProviderKey]

		if !found {
			provider, err := k.GetFinalityProvider(ctx, delegation.FpBtcPkList[0].MustMarshal())
			if err != nil {
				panic(err)
			}

			if provider.IsSlashed() {
				fpData = &FpInfo{
					distInfo: nil,
					slashed:  true,
				}
				finalityProviders[finalityProviderKey] = fpData
				return true
			}

			distInfo := types.NewFinalityProviderDistInfo(provider)

			distInfo.AddBTCDel(delegation, btcTipHeight, wValue, covenantQuorum)

			finalityProviders[finalityProviderKey] = &FpInfo{
				distInfo: distInfo,
				slashed:  false,
			}

			return true
		}

		if fpData.slashed {
			return true
		}

		finalityProviders[finalityProviderKey].distInfo.AddBTCDel(delegation, btcTipHeight, wValue, covenantQuorum)

		return true
	})

	var provSlice []*types.FinalityProviderDistInfo

	rdc := types.NewRewardDistCache()

	for _, fpData := range finalityProviders {
		fpCopy := fpData

		if fpCopy.slashed {
			continue
		}

		if fpCopy.distInfo.TotalVotingPower == 0 {
			continue
		}
		provSlice = append(provSlice, fpCopy.distInfo)
	}

	var maxValidators = int(k.GetParams(ctx).MaxActiveFinalityProviders)

	var activeValidators int

	if len(provSlice) >= maxValidators {
		activeValidators = maxValidators
	} else {
		activeValidators = len(provSlice)
	}

	sort.SliceStable(provSlice, func(i, j int) bool {
		return provSlice[i].TotalVotingPower > provSlice[j].TotalVotingPower
	})

	babylonTipHeight := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)

	for i := 0; i < activeValidators; i++ {
		k.SetVotingPower(ctx, provSlice[i].BtcPk.MustMarshal(), babylonTipHeight, provSlice[i].TotalVotingPower)
		rdc.AddFinalityProviderDistInfo(provSlice[i])
	}

	k.setRewardDistCache(ctx, babylonTipHeight, rdc)
	return nil
}

// new consensus compatible BeginBlocker
func (k Keeper) BeginBlocker(ctx context.Context) error {
	k.IndexBTCHeight(ctx)

	covenantQuorum := k.GetParams(ctx).CovenantQuorum
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	distInfos := make(map[chainhash.Hash]*types.BTCDelDistInfo)

	k.IterateBTCDelsKeys(ctx, func(key *chainhash.Hash, btcDel *types.BTCDelegation) bool {
		distInfo := &types.BTCDelDistInfo{
			BabylonPk:   btcDel.BabylonPk,
			VotingPower: btcDel.VotingPower(btcTipHeight, wValue, covenantQuorum),
		}
		if distInfo.VotingPower > 0 {
			distInfos[*key] = distInfo
		}
		return true
	})

	activeFps := []*types.FinalityProviderWithMeta{}

	rdc := types.NewRewardDistCache()

	k.IterateActiveFPs(
		ctx,
		func(fp *types.FinalityProvider) bool {
			fpDistInfo := types.NewFinalityProviderDistInfo(fp)

			k.IterateBTCDelegationsHashes(ctx, fp.BtcPk, func(hash *chainhash.Hash) bool {
				distInfo, found := distInfos[*hash]

				if !found {
					return true
				}

				fpDistInfo.AddBTCDistInfo(distInfo)
				return true
			})

			if fpDistInfo.TotalVotingPower > 0 {
				activeFP := &types.FinalityProviderWithMeta{
					BtcPk:       fp.BtcPk,
					VotingPower: fpDistInfo.TotalVotingPower,
				}
				activeFps = append(activeFps, activeFP)
				rdc.AddFinalityProviderDistInfo(fpDistInfo)
			}

			return true
		},
	)

	// filter out top `MaxActiveFinalityProviders` active finality providers in terms of voting power
	activeFps = types.FilterTopNFinalityProviders(activeFps, k.GetParams(ctx).MaxActiveFinalityProviders)
	// set voting power table
	babylonTipHeight := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	for _, fp := range activeFps {
		k.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), babylonTipHeight, fp.VotingPower)
	}

	// set the reward distribution cache of the current height
	// TODO: only give rewards to top N finality providers and their BTC delegations
	k.setRewardDistCache(ctx, uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height), rdc)

	return nil
}
