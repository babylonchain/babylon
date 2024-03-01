package keeper

import (
	"context"
	"fmt"

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
func (k Keeper) BeginBlocker(ctx context.Context) error {
	k.IndexBTCHeight(ctx)

	covenantQuorum := k.GetParams(ctx).CovenantQuorum
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	distInfos := make(map[chainhash.Hash]*types.BTCDelDistInfo)

	k.IterateBTCDelsKeys(ctx, func(key chainhash.Hash, btcDel *types.BTCDelegation) bool {
		distInfo := &types.BTCDelDistInfo{
			BabylonPk:   btcDel.BabylonPk,
			VotingPower: btcDel.VotingPower(btcTipHeight, wValue, covenantQuorum),
		}
		if distInfo.VotingPower > 0 {
			distInfos[key] = distInfo
		}
		return true
	})

	activeFps := []*types.FinalityProviderWithMeta{}

	rdc := types.NewRewardDistCache()

	k.IterateActiveFPs(
		ctx,
		func(fp *types.FinalityProvider) bool {
			fpDistInfo := types.NewFinalityProviderDistInfo(fp)

			k.IterateBTCDelegationsHashes(ctx, fp.BtcPk, func(hash chainhash.Hash) bool {
				distInfo, found := distInfos[hash]

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
