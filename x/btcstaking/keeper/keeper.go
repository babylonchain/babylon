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
	// index BTC height at the current height
	k.IndexBTCHeight(ctx)

	params := k.GetParams(ctx)
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

	// prepare metrics for {active, inactive} finality providers,
	// {pending, active, unbonded BTC delegations}, and total staked Bitcoins
	// NOTE: slashed finality providers and BTC delegations are recorded upon
	// slashing events rather than here
	var (
		totalFPsMap          = map[string]struct{}{}
		numStakedSats uint64 = 0
		numDelsMap           = map[types.BTCDelegationStatus]int{
			types.BTCDelegationStatus_PENDING:  0,
			types.BTCDelegationStatus_ACTIVE:   0,
			types.BTCDelegationStatus_UNBONDED: 0,
		}
	)

	k.IterateActiveFPsAndBTCDelegations(
		ctx,
		func(fp *types.FinalityProvider, btcDel *types.BTCDelegation) bool {
			fpBTCPKHex := fp.BtcPk.MarshalHex()

			// record this finality provider
			totalFPsMap[fpBTCPKHex] = struct{}{}

			// record active finality providers
			power := btcDel.VotingPower(btcTipHeight, wValue, params.CovenantQuorum)
			if power == 0 {
				return true // skip if no voting power
			}
			fpPowerMap[fpBTCPKHex] += power

			// record metrics
			numStakedSats += power
			numDelsMap[btcDel.GetStatus(btcTipHeight, wValue, params.CovenantQuorum)]++

			// create fp dist info if not exist
			if _, ok := fpDistMap[fpBTCPKHex]; !ok {
				fpDistMap[fpBTCPKHex] = types.NewFinalityProviderDistInfo(fp)
			}
			// append BTC delegation
			fpDistMap[fpBTCPKHex].AddBTCDel(btcDel, btcTipHeight, wValue, params.CovenantQuorum)
			return true
		},
	)

	// record metrics for finality providers and total staked BTCs
	numFPs := len(totalFPsMap)
	numActiveFPs := numFPs
	if numActiveFPs > int(params.MaxActiveFinalityProviders) {
		numActiveFPs = int(params.MaxActiveFinalityProviders)
	}
	types.RecordActiveFinalityProviders(numActiveFPs)
	types.RecordInactiveFinalityProviders(numFPs - numActiveFPs)
	numStakedBTCs := float32(numStakedSats / SatoshisPerBTC)
	types.RecordMetricsKeyStakedBitcoins(numStakedBTCs)
	// record metrics for BTC delegations (NOTE: slashed BTC delegations are recorded upon slashing)
	for status, num := range numDelsMap {
		types.RecordBTCDelegations(num, status)
	}

	// return directly if there is no active finality provider
	if len(fpPowerMap) == 0 {
		return nil
	}

	// get top N finality providers and set their voting power to KV store
	k.setCurrentTopNVotingPower(ctx, fpPowerMap)

	// create reward distribution cache
	rdc := types.NewRewardDistCache()
	for fpBTCPKHex := range fpDistMap {
		// try to add this finality provider distribution info to reward distribution cache
		rdc.AddFinalityProviderDistInfo(fpDistMap[fpBTCPKHex])
	}

	// all good, set the reward distribution cache of the current height
	k.setCurrentRewardDistCache(ctx, rdc)

	return nil
}
