package keeper

import (
	"context"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	fps, err := k.finalityProviders(ctx)
	if err != nil {
		return nil, err
	}

	dels, err := k.getBTCDelegations(ctx)
	if err != nil {
		return nil, err
	}

	vpFps, err := k.fpVotingPowers(ctx)
	if err != nil {
		return nil, err
	}

	btcDels, err := k.getBTCDelegators(ctx)
	if err != nil {
		return nil, err
	}

	evts, err := k.getEventIdxs(ctx)
	if err != nil {
		return nil, err
	}

	vpsCache, err := k.votingPowersDistCacheBlkHeight(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		FinalityProviders: fps,
		BtcDelegations:    dels,
		VotingPowers:      vpFps,
		BlockHeightChains: k.blockHeightChains(ctx),
		BtcDelegators:     btcDels,
		Events:            evts,
		VpDstCache:        vpsCache,
	}, nil
}

func (k Keeper) finalityProviders(ctx context.Context) ([]*types.FinalityProvider, error) {
	fps := make([]*types.FinalityProvider, 0)
	iter := k.finalityProviderStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var fp types.FinalityProvider
		if err := fp.Unmarshal(iter.Key()); err != nil {
			return nil, err
		}
		fps = append(fps, &fp)
	}

	return fps, nil
}

func (k Keeper) getBTCDelegations(ctx context.Context) ([]*types.BTCDelegation, error) {
	dels := make([]*types.BTCDelegation, 0)
	iter := k.btcDelegationStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var del types.BTCDelegation
		if err := del.Unmarshal(iter.Key()); err != nil {
			return nil, err
		}
		dels = append(dels, &del)
	}

	return dels, nil
}

// fpVotingPowers gets the voting power of a given finality provider at a given Babylon height.
func (k Keeper) fpVotingPowers(ctx context.Context) ([]*types.VotingPowerFP, error) {
	iter := k.votingPowerStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	vpFps := make([]*types.VotingPowerFP, 0)

	for ; iter.Valid(); iter.Next() {
		blkHeight, fpBTCPK, err := bbn.ParseBlkHeightAndPubKeyFromStoreKey(iter.Key())
		if err != nil {
			return nil, err
		}

		vp := sdk.BigEndianToUint64(iter.Value())
		vpFps = append(vpFps, &types.VotingPowerFP{
			BlockHeight: blkHeight,
			FpBtcPk:     fpBTCPK,
			VotingPower: vp,
		})
	}

	return vpFps, nil
}

func (k Keeper) blockHeightChains(ctx context.Context) []*types.BlockHeightBbnToBtc {
	iter := k.btcHeightStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	blocks := make([]*types.BlockHeightBbnToBtc, 0)
	for ; iter.Valid(); iter.Next() {
		blocks = append(blocks, &types.BlockHeightBbnToBtc{
			BlockHeightBbn: sdk.BigEndianToUint64(iter.Key()),
			BlockHeightBtc: sdk.BigEndianToUint64(iter.Value()),
		})
	}

	return blocks
}

func (k Keeper) getBTCDelegators(ctx context.Context) ([]*types.BTCDelegator, error) {
	iter := k.btcDelegationStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	dels := make([]*types.BTCDelegator, 0)
	for ; iter.Valid(); iter.Next() {
		fpBTCPK, delBTCPK, err := bbn.ParseBIP340PubKeysFromStoreKey(iter.Key())
		if err != nil {
			return nil, err
		}
		var btcDelIndex types.BTCDelegatorDelegationIndex
		if err := btcDelIndex.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}

		dels = append(dels, &types.BTCDelegator{
			Idx:      &btcDelIndex,
			FpBtcPk:  fpBTCPK,
			DelBtcPk: delBTCPK,
		})
	}

	return dels, nil
}

// getEventIdxs sets an event into the store.
func (k Keeper) getEventIdxs(
	ctx context.Context,
) ([]*types.EventIndex, error) {
	iter := k.powerDistUpdateEventStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	evts := make([]*types.EventIndex, 0)
	for ; iter.Valid(); iter.Next() {
		blkHeight, idx, err := bbn.ParseUintsFromStoreKey(iter.Key())
		if err != nil {
			return nil, err
		}

		var evt types.EventPowerDistUpdate
		if err := evt.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}

		evts = append(evts, &types.EventIndex{
			Idx:            idx,
			BlockHeightBtc: blkHeight,
			Event:          &evt,
		})
	}

	return evts, nil
}

func (k Keeper) votingPowersDistCacheBlkHeight(ctx context.Context) ([]*types.VotingPowerDistCacheBlkHeight, error) {
	vps := make([]*types.VotingPowerDistCacheBlkHeight, 0)
	iter := k.votingPowerDistCacheStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var dc types.VotingPowerDistCache
		if err := dc.Unmarshal(iter.Key()); err != nil {
			return nil, err
		}
		vps = append(vps, &types.VotingPowerDistCacheBlkHeight{
			BlockHeight:    sdk.BigEndianToUint64(iter.Key()),
			VpDistribution: &dc,
		})
	}

	return vps, nil
}
