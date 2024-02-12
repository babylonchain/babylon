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

const (
	SatoshisPerBTC = 100_000_000
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

	params := k.GetParams(ctx)
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	distInfos := make(map[chainhash.Hash]*types.BTCDelDistInfo)

	var numDelsMap = map[types.BTCDelegationStatus]int{
		types.BTCDelegationStatus_PENDING:  0,
		types.BTCDelegationStatus_ACTIVE:   0,
		types.BTCDelegationStatus_UNBONDED: 0,
	}
	k.IterateBTCDelsKeys(ctx, func(key chainhash.Hash, btcDel *types.BTCDelegation) bool {
		distInfo := &types.BTCDelDistInfo{
			BabylonPk:   btcDel.BabylonPk,
			VotingPower: btcDel.VotingPower(btcTipHeight, wValue, params.CovenantQuorum),
		}
		numDelsMap[btcDel.GetStatus(btcTipHeight, wValue, params.CovenantQuorum)]++

		if distInfo.VotingPower > 0 {
			distInfos[key] = distInfo
		}
		return true
	})

	activeFps := []*types.FinalityProviderWithMeta{}
	// prepare for recording finality providers and their BTC delegations
	// for rewards
	dc := types.NewRewardDistCache()

	// prepare metrics for {active, inactive} finality providers,
	// {pending, active, unbonded BTC delegations}, and total staked Bitcoins
	// NOTE: slashed finality providers and BTC delegations are recorded upon
	// slashing events rather than here
	var (
		numFPs        int    = 0
		numStakedSats uint64 = 0
	)

	k.IterateActiveFPs(
		ctx,
		func(fp *types.FinalityProvider) bool {
			fpDistInfo := types.NewFinalityProviderDistInfo(fp)

			k.IterateBTCDelegationsHashes(ctx, fp.BtcPk, func(hash chainhash.Hash) bool {
				distInfo, found := distInfos[hash]

				if !found {
					return true
				}

				numStakedSats += distInfo.VotingPower

				fpDistInfo.AddBTCDistInfo(distInfo)
				return true
			})

			if fpDistInfo.TotalVotingPower > 0 {
				activeFP := &types.FinalityProviderWithMeta{
					BtcPk:       fp.BtcPk,
					VotingPower: fpDistInfo.TotalVotingPower,
				}
				activeFps = append(activeFps, activeFP)
				dc.AddFinalityProviderDistInfo(fpDistInfo)
			}

			return true
		},
	)
	// record metrics for finality providers and total staked BTCs
	numActiveFPs := min(numFPs, int(params.MaxActiveFinalityProviders))
	types.RecordActiveFinalityProviders(numActiveFPs)
	types.RecordInactiveFinalityProviders(numFPs - numActiveFPs)
	numStakedBTCs := float32(numStakedSats / SatoshisPerBTC)
	types.RecordMetricsKeyStakedBitcoins(numStakedBTCs)
	// record metrics for BTC delegations
	for status, num := range numDelsMap {
		types.RecordBTCDelegations(num, status)
	}

	// filter out top `MaxActiveFinalityProviders` active finality providers in terms of voting power
	activeFps = types.FilterTopNFinalityProviders(activeFps, params.MaxActiveFinalityProviders)
	// set voting power table
	babylonTipHeight := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	for _, fp := range activeFps {
		k.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), babylonTipHeight, fp.VotingPower)
	}

	// set the reward distribution cache of the current height
	// TODO: only give rewards to top N finality providers and their BTC delegations
	k.setRewardDistCache(ctx, babylonTipHeight, dc)

	return nil
}
