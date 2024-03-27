package keeper

import (
	"context"
	"fmt"

	btcstk "github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	// save all past params versions
	for _, p := range gs.Params {
		params := p
		if err := k.SetParams(ctx, *params); err != nil {
			return err
		}
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

	// Events are generated on block `N` to be processed at block `N+1`
	// When ExportGenesis is called the node already stopped at block N.
	// In this case the events on the state would refer to the block `N+1`
	// Since InitGenesis occurs before BeginBlock, the genesis state would be properly
	// stored in the KV store for when BeginBlock process the events.
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

	dels, err := k.btcDelegations(ctx)
	if err != nil {
		return nil, err
	}

	vpFps, err := k.fpVotingPowers(ctx)
	if err != nil {
		return nil, err
	}

	btcDels, err := k.btcDelegators(ctx)
	if err != nil {
		return nil, err
	}

	evts, err := k.eventIdxs(ctx)
	if err != nil {
		return nil, err
	}

	vpsCache, err := k.votingPowersDistCacheBlkHeight(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GenesisState{
		Params:            k.GetAllParams(ctx),
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
		if err := fp.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}
		fps = append(fps, &fp)
	}

	return fps, nil
}

func (k Keeper) btcDelegations(ctx context.Context) ([]*types.BTCDelegation, error) {
	dels := make([]*types.BTCDelegation, 0)
	iter := k.btcDelegationStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var del types.BTCDelegation
		if err := del.Unmarshal(iter.Value()); err != nil {
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
		blkHeight, fpBTCPK, err := btcstk.ParseBlkHeightAndPubKeyFromStoreKey(iter.Key())
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

func (k Keeper) btcDelegators(ctx context.Context) ([]*types.BTCDelegator, error) {
	iter := k.btcDelegatorStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	dels := make([]*types.BTCDelegator, 0)
	for ; iter.Valid(); iter.Next() {
		fpBTCPK, delBTCPK, err := parseBIP340PubKeysFromStoreKey(iter.Key())
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

// eventIdxs sets an event into the store.
func (k Keeper) eventIdxs(
	ctx context.Context,
) ([]*types.EventIndex, error) {
	iter := k.powerDistUpdateEventStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	evts := make([]*types.EventIndex, 0)
	for ; iter.Valid(); iter.Next() {
		blkHeight, idx, err := parseUintsFromStoreKey(iter.Key())
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
		if err := dc.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}
		vps = append(vps, &types.VotingPowerDistCacheBlkHeight{
			BlockHeight:    sdk.BigEndianToUint64(iter.Key()),
			VpDistribution: &dc,
		})
	}

	return vps, nil
}

func (k Keeper) setBlockHeightChains(ctx context.Context, blocks *types.BlockHeightBbnToBtc) {
	store := k.btcHeightStore(ctx)
	store.Set(sdk.Uint64ToBigEndian(blocks.BlockHeightBbn), sdk.Uint64ToBigEndian(blocks.BlockHeightBtc))
}

// setEventIdx sets an event into the store.
func (k Keeper) setEventIdx(
	ctx context.Context,
	evt *types.EventIndex,
) error {
	store := k.powerDistUpdateEventBtcHeightStore(ctx, evt.BlockHeightBtc)

	bz, err := evt.Event.Marshal()
	if err != nil {
		return err
	}
	store.Set(sdk.Uint64ToBigEndian(evt.Idx), bz)

	return nil
}

// parseUintsFromStoreKey expects to receive a key with
// BigEndianUint64(blkHeight) || BigEndianUint64(Idx)
func parseUintsFromStoreKey(key []byte) (blkHeight, idx uint64, err error) {
	sizeBigEndian := 8
	if len(key) < sizeBigEndian*2 {
		return 0, 0, fmt.Errorf("key not long enough to parse two uint64: %s", key)
	}

	return sdk.BigEndianToUint64(key[:sizeBigEndian]), sdk.BigEndianToUint64(key[sizeBigEndian:]), nil
}

// parseBIP340PubKeysFromStoreKey expects to receive a key with
// BIP340PubKey(fpBTCPK) || BIP340PubKey(delBTCPK)
func parseBIP340PubKeysFromStoreKey(key []byte) (fpBTCPK, delBTCPK *bbn.BIP340PubKey, err error) {
	if len(key) < bbn.BIP340PubKeyLen*2 {
		return nil, nil, fmt.Errorf("key not long enough to parse two BIP340PubKey: %s", key)
	}

	fpBTCPK, err = bbn.NewBIP340PubKey(key[:bbn.BIP340PubKeyLen])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse pub key from key %w: %w", bbn.ErrUnmarshal, err)
	}

	delBTCPK, err = bbn.NewBIP340PubKey(key[bbn.BIP340PubKeyLen:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse pub key from key %w: %w", bbn.ErrUnmarshal, err)
	}

	return fpBTCPK, delBTCPK, nil
}
