package keeper

import (
	"context"

	"github.com/babylonchain/babylon/x/btcstaking/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}

	for _, fp := range gs.FinalityProviders {
		k.SetFinalityProvider(ctx, fp)
	}

	for _, btcDel := range gs.BtcDelegations {
		k.setBTCDelegation(ctx, btcDel)
	}

	for _, fpVP := range gs.VotingPowers {
		k.SetVotingPower(ctx, *fpVP.FpBtcPk, fpVP.BlockHeight, fpVP.VotingPower)
	}

	for _, blocks := range gs.BlockHeightChains {
		k.setBlockHeightChains(ctx, blocks)
	}

	for _, del := range gs.BtcDelegators {
		k.setBTCDelegatorDelegationIndex(ctx, del.FpBtcPk, del.DelBtcPk, del.Idx)
	}
	for _, evt := range gs.Events {
		if err := k.setEventIdx(ctx, evt); err != nil {
			return err
		}
	}
	for _, vpCache := range gs.VpDstCache {
		k.setVotingPowerDistCache(ctx, vpCache.BlockHeight, vpCache.VpDistribution)
	}

	return nil
}
