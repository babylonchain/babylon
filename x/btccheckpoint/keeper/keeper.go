package keeper

import (
	"fmt"

	"math/big"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc                  codec.BinaryCodec
		storeKey             sdk.StoreKey
		memKey               sdk.StoreKey
		paramstore           paramtypes.Subspace
		btcLightClientKeeper types.BTCLightClientKeeper
		checkpointingKeeper  types.CheckpointingKeeper
		kDeep                uint64
		wDeep                uint64
		powLimit             *big.Int
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	bk types.BTCLightClientKeeper,
	ck types.CheckpointingKeeper,
	// Those are node level constants should go to some kind of global node config
	kDeep uint64,
	wDeep uint64,
	powLimit *big.Int,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                  cdc,
		storeKey:             storeKey,
		memKey:               memKey,
		paramstore:           ps,
		btcLightClientKeeper: bk,
		checkpointingKeeper:  ck,
		kDeep:                kDeep,
		wDeep:                wDeep,
		powLimit:             powLimit,
	}
}

func (k Keeper) GetPowLimit() *big.Int {
	return k.powLimit
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetBlockHeight(ctx sdk.Context, b *bbn.BTCHeaderHashBytes) (uint64, error) {
	return k.btcLightClientKeeper.BlockHeight(ctx, b)
}

func (k Keeper) CheckHeaderIsKnown(ctx sdk.Context, hash *bbn.BTCHeaderHashBytes) bool {
	_, err := k.btcLightClientKeeper.MainChainDepth(ctx, hash)
	return err == nil
}

func (k Keeper) MainChainDepth(ctx sdk.Context, hash *bbn.BTCHeaderHashBytes) (uint64, bool, error) {
	depth, err := k.btcLightClientKeeper.MainChainDepth(ctx, hash)

	if err != nil {
		return 0, false, err
	}

	if depth < 0 {
		return 0, false, nil
	}

	return uint64(depth), true, nil
}

func (k Keeper) IsAncestor(ctx sdk.Context, parentHash *bbn.BTCHeaderHashBytes, childHash *bbn.BTCHeaderHashBytes) (bool, error) {
	return k.btcLightClientKeeper.IsAncestor(ctx, parentHash, childHash)
}

func (k Keeper) GetCheckpointEpoch(ctx sdk.Context, c []byte) (uint64, error) {
	return k.checkpointingKeeper.CheckpointEpoch(ctx, c)
}

func (k Keeper) SubmissionExists(ctx sdk.Context, sk types.SubmissionKey) bool {
	store := ctx.KVStore(k.storeKey)
	kBytes := k.cdc.MustMarshal(&sk)
	return store.Has(kBytes)
}

// Return epoch data for given epoch, if there is not epoch data yet returns nil
func (k Keeper) GetEpochData(ctx sdk.Context, e uint64) *types.EpochData {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.GetEpochIndexKey(e))

	if len(bytes) == 0 {
		return nil
	}

	ed := &types.EpochData{}
	k.cdc.MustUnmarshal(bytes, ed)
	return ed
}

func (k Keeper) saveEpochData(ctx sdk.Context, e uint64, ed *types.EpochData) {
	store := ctx.KVStore(k.storeKey)
	ek := types.GetEpochIndexKey(e)
	eb := k.cdc.MustMarshal(ed)
	store.Set(ek, eb)
}

func (k Keeper) AddEpochSubmission(
	ctx sdk.Context,
	epochNum uint64,
	sk types.SubmissionKey,
	sd types.SubmissionData,
	epochRawCheckpoint []byte) error {

	ed := k.GetEpochData(ctx, epochNum)

	// TODO: SaveEpochData and SaveSubmission should be done in one transaction.
	// Not sure cosmos-sdk has facialities to do it.
	// Otherwise it is possible to end up with node which updated submission list
	// but did not save submission itself.

	// if ed is nil, it means it is our first submission for this epoch
	if ed == nil {
		// we do not have any data saved yet
		newEd := types.NewEmptyEpochData(epochRawCheckpoint)
		ed = &newEd
	}

	if ed.Status == types.Confirmed || ed.Status == types.Finalized {
		// we already received submission which confirmed/finalized epoch, there is no
		// need of accepting any more
		return types.ErrEpochAlreadyConfirmedOrFinalized
	}

	onMainChain, err := k.checkSubmissionOnMainChain(ctx, sk)

	if err != nil {
		panic("All submissions should be known to light client during submission addition")
	}

	if ed.Status == types.Signed && onMainChain {
		// It is first epoch submission which is on the main chain, inform checkpointing module
		// about it and change epoch status to submited
		ed.Status = types.Submitted
		k.checkpointingKeeper.SetCheckpointSubmitted(ctx, epochNum)
	}

	ed.AppendKey(sk)
	// Even though it is submission which is on fork rather that main chain it still
	// counts as unconfirmed submission.
	k.addToUnconfirmed(ctx, sk)
	k.saveEpochData(ctx, epochNum, ed)
	k.saveSubmission(ctx, sk, sd)
	return nil
}

func (k Keeper) addToUnconfirmed(ctx sdk.Context, sk types.SubmissionKey) {
	store := ctx.KVStore(k.storeKey)
	uk := types.UnconfiredSubmissionsKey(k.cdc, &sk)
	v := k.cdc.MustMarshal(&sk)
	store.Set(uk, v)
}

func (k Keeper) saveSubmission(ctx sdk.Context, sk types.SubmissionKey, sd types.SubmissionData) {
	store := ctx.KVStore(k.storeKey)
	kBytes := types.PrefixedSubmisionKey(k.cdc, &sk)
	sBytes := k.cdc.MustMarshal(&sd)
	store.Set(kBytes, sBytes)
}

// GetSubmissionData return submission data for given key, return nil if there is not data
// under givem key
func (k Keeper) GetSubmissionData(ctx sdk.Context, sk types.SubmissionKey) *types.SubmissionData {
	store := ctx.KVStore(k.storeKey)
	kBytes := types.PrefixedSubmisionKey(k.cdc, &sk)
	sdBytes := store.Get(kBytes)

	if len(sdBytes) == 0 {
		return nil
	}

	var sd types.SubmissionData
	k.cdc.MustUnmarshal(sdBytes, &sd)
	return &sd
}

// getSubmissionDataExists retrive submissions data, panics if data does not exists
// should only be called when data for sure is in store
func (k Keeper) getSubmissionDataExists(ctx sdk.Context, sk types.SubmissionKey) types.SubmissionData {
	store := ctx.KVStore(k.storeKey)
	kBytes := types.PrefixedSubmisionKey(k.cdc, &sk)
	sdBytes := store.Get(kBytes)
	if sdBytes == nil {
		panic("Submission data should exists in the store")
	}

	var sd types.SubmissionData
	k.cdc.MustUnmarshal(sdBytes, &sd)
	return sd
}

// checkSubmissionNDeepOnMainChain checks if all headers from submission key are one the main
// chain (which also guarantess that they have ancestry relationship or are the same header)
// and are at least n blocks deep
// returns error if at least one of the headers is no longer known to light client
func (k Keeper) checkSubmissionNDeepOnMainChain(ctx sdk.Context, sk types.SubmissionKey, n uint64) (bool, bool, error) {
	var onMain bool = true
	var allAtLeastNDeep = true
	for _, tk := range sk.Key {
		depth, onMainChain, e := k.MainChainDepth(ctx, tk.Hash)
		if e != nil {
			return false, false, e
		}

		if !onMainChain {
			onMain = false
		}

		if depth < n {
			allAtLeastNDeep = false
		}
	}
	return onMain, allAtLeastNDeep, nil
}

func (k Keeper) checkSubmissionOnMainChain(ctx sdk.Context, sk types.SubmissionKey) (bool, error) {
	var onMain bool = true
	for _, tk := range sk.Key {
		_, onMainChain, e := k.MainChainDepth(ctx, tk.Hash)

		if e != nil {
			return false, e
		}

		if !onMainChain {
			onMain = false
		}
	}
	return onMain, nil
}

func (k Keeper) checkSubmissionConfirmed(ctx sdk.Context, sk types.SubmissionKey) (bool, bool, error) {
	return k.checkSubmissionNDeepOnMainChain(ctx, sk, k.kDeep)
}

func (k Keeper) checkSubmissionFinlized(ctx sdk.Context, sk types.SubmissionKey) (bool, bool, error) {
	return k.checkSubmissionNDeepOnMainChain(ctx, sk, k.wDeep)
}

func (k Keeper) promoteUnconfirmedToConfirmed(ctx sdk.Context, sk types.SubmissionKey) {
	store := ctx.KVStore(k.storeKey)
	subKey := k.cdc.MustMarshal(&sk)
	unconfirmedKey := types.UnconfiredSubmissionsKey(k.cdc, &sk)
	confirmedKey := types.ConfirmedSubmissionsKey(k.cdc, &sk)

	// Promotion of submision from submitted state to confirmed state is just
	// - deleting unconfirmed index
	// - saving confirmed index
	store.Delete(unconfirmedKey)
	store.Set(confirmedKey, subKey)
}

func (k Keeper) promoteConfirmedToFinalized(ctx sdk.Context, sk types.SubmissionKey) {
	store := ctx.KVStore(k.storeKey)
	subKey := k.cdc.MustMarshal(&sk)
	confirmedKey := types.ConfirmedSubmissionsKey(k.cdc, &sk)
	finalizedKey := types.FinalizedSubmissionsKey(k.cdc, &sk)

	// Promotion of submision from submitted state to confirmed state is just
	// - deleting confirmed index
	// - saving finalized index
	store.Delete(confirmedKey)
	store.Set(finalizedKey, subKey)
}

// Iterate over all unconfirmed submissions, and check their bitcoin status
// Note: Alternative to iterator would be separate key value pair in db holding
// list of all unconfirmed submission keys.
// Approach with iterator was taken as:
// - There can be many unconfirmed submissions across many epochs
// - pruning is a bit more streight forward with iterator apporoach
func (k Keeper) checkUnconfirmed(ctx sdk.Context) {

	newConfirmed := []types.SubmissionKey{}

	store := ctx.KVStore(k.storeKey)

	// iterator over all unconfirmed submissions
	iterator := sdk.KVStorePrefixIterator(store, types.UnconfirmedIndexPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		skBytes := iterator.Value()
		var sk types.SubmissionKey
		k.cdc.MustUnmarshal(skBytes, &sk)

		onMainChain, deepEnough, err := k.checkSubmissionConfirmed(ctx, sk)

		if err != nil {
			// submission which was known to lighclient is no longer known
			// TODO Decide how to best handle submissions forgotten by light client
			// There are two main options:
			// 1. Delete them - probably best in terms of space usage, but a lot of indexes
			// need updating when deleting submissions.
			// 2. Mark them as forgotten - they would linger a bit longer but the only
			// index needed updaing in Submitted
			// Whatever route will be taken, one need to remeber to inform checkpointing
			// module if all checkpoints from epoch will be lost
			continue
		}

		// TODO Add handling of the case when onMainChain is false, which requires checking
		// state of this submission epoch data, and if its Submitted it means that,
		// submission which was on main chain suddenly became orphaned. If all submissions
		// of the epoch become orphaned we need to inform checkpoinitng module about it

		if onMainChain && deepEnough {
			// we have new confirmed submission
			newConfirmed = append(newConfirmed, sk)
		}
	}

	if len(newConfirmed) == 0 {
		// no new confirmed sumbmissions
		return
	}

	newConfirmedEpochs := map[uint64]bool{}
	for _, newConfirmedSubKey := range newConfirmed {
		// if we would not have submission under this key, then something is really wrong
		// with our data model
		sd := k.getSubmissionDataExists(ctx, newConfirmedSubKey)

		_, alreadyConfirmed := newConfirmedEpochs[sd.Epoch]

		if alreadyConfirmed {
			// one of the earlier newConfirmed submission keys already confirmed this epoch
			// and we already processed
			continue
		}

		// we need to check if this is first confirmed submission in given epoch
		ed := k.GetEpochData(ctx, sd.Epoch)

		if ed == nil {
			// if we do not have any data about epoch, something is really wrong with
			// data model
			panic("Submission without existing epoch")
		}

		if ed.Status != types.Submitted {
			// epoch is already finalized/confirmed, no need to do any thing else as other
			// submission confirmed/finalized this epoch
			continue
		}

		if len(ed.Key) == 0 {
			// this check is here only to check data model consistency. It should probably
			// be hidden behind some debug compile flag (not sure golang has such things)
			panic("Broken data model. There should be at least one submmission on the list of epoch submissions")
		}

		// at this point we know that:
		// - one of submitted submissions changed state from submitted to confirmed
		// - epoch for this submission is not yet confirmed or finalized which means
		// there aren't any confirmed finalized submission for this epoch ye
		// we need to check if there are any other submission in this epoch which
		// changed its state
		newConfirmedEpochs[sd.Epoch] = true
		ed.Status = types.Confirmed

		// Save epoch with confirmed status and infrom checkpointing module about
		// new confirmed checpoint
		k.saveEpochData(ctx, sd.Epoch, ed)
		k.checkpointingKeeper.SetCheckpointConfirmed(ctx, sd.Epoch)

		// TODO Rewards.
		// 1. Check if any other submission from this epoch did not become confirmed
		// 2. If yes, determine the better one of the newly confirmed submissions
		// Definition of `better` need some tought. Shoud those be firstly submitted
		// submission, or maybe submission which is oldest on btc chain  ?
		// 3. Pay the reward to the best submission. It will probably require mint
		// keeper os smth like that
	}

	for _, newConfirmedSubKey := range newConfirmed {
		// Promote all newly confirmed keys
		// It could be done in loop which handles epoch but it is a bit cleaner that way
		// this will be especially clear when working on rewards
		k.promoteUnconfirmedToConfirmed(ctx, newConfirmedSubKey)
	}

}

func (k Keeper) checkConfirmed(ctx sdk.Context) {

	newFinalized := []types.SubmissionKey{}

	store := ctx.KVStore(k.storeKey)
	// iterator over all already confirmed submissions
	iterator := sdk.KVStorePrefixIterator(store, types.ConfirmedIndexPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		skBytes := iterator.Value()
		var sk types.SubmissionKey
		k.cdc.MustUnmarshal(skBytes, &sk)

		onMainChain, deepEnough, err := k.checkSubmissionFinlized(ctx, sk)

		if err != nil {
			// Previously confirmed submission is now unknown to btc light client.
			// Most probably it means that chain grown over the kept history.
			// TODO. Decide what to do in that case. Most probably when pruning of old
			// epoch data is implemented and all parameters of btclightclient and
			// btccheckpoint are correct it should not happen Then we can panic here
			// for now ignore
			//
			continue
		}

		// TODO consider if we should check if submission ended on fork. It would mean
		// that something fishy is going on or we have some kind of bug in light client

		if onMainChain && deepEnough {
			newFinalized = append(newFinalized, sk)
		}
	}

	if len(newFinalized) == 0 {
		return
	}

	newFinalizedEpochs := map[uint64]bool{}
	for _, newFinalizedSubKey := range newFinalized {
		// if we would not have submission under this key, then something is really wrong
		// with our data model
		sd := k.getSubmissionDataExists(ctx, newFinalizedSubKey)

		_, alreadyFinalized := newFinalizedEpochs[sd.Epoch]

		if alreadyFinalized {
			// one of the earlier newConfirmed submission keys already confirmed this epoch
			// and we already processed
			continue
		}

		// we need to check if this is first finalized submission in given epoch
		ed := k.GetEpochData(ctx, sd.Epoch)

		if ed == nil {
			// if we do not have any data about epoch, something is really wrong with
			// data model
			panic("Submission without existing epoch")
		}

		if ed.Status == types.Submitted {
			panic("Got confired submission for not confirmed epoch")
		}

		if ed.Status == types.Finalized {
			// epoch already finalized no need to do anything
			continue
		}

		// at this point:
		// - we have new finalized submission for confirmed epoch
		// so:
		// - save epoch data with new state
		// - inform checkpointing about it
		newFinalizedEpochs[sd.Epoch] = true
		ed.Status = types.Finalized
		k.saveEpochData(ctx, sd.Epoch, ed)
		k.checkpointingKeeper.SetCheckpointFinalized(ctx, sd.Epoch)

		// TODO Consider how to prune submissions
	}

	for _, newFinalizedSubKey := range newFinalized {
		k.promoteConfirmedToFinalized(ctx, newFinalizedSubKey)
	}

}

func (k Keeper) getSubmissionsWithPrefix(ctx sdk.Context, prefix []byte) []types.SubmissionKey {
	sumbmissionKeys := []types.SubmissionKey{}

	store := ctx.KVStore(k.storeKey)

	// iterator over all submissions  with prefix
	iterator := sdk.KVStorePrefixIterator(store, prefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		skBytes := iterator.Value()
		var sk types.SubmissionKey
		k.cdc.MustUnmarshal(skBytes, &sk)
		sumbmissionKeys = append(sumbmissionKeys, sk)
	}
	return sumbmissionKeys
}

func (k Keeper) GetAllUnconfirmedSubmissions(ctx sdk.Context) []types.SubmissionKey {
	return k.getSubmissionsWithPrefix(ctx, types.UnconfirmedIndexPrefix)
}

func (k Keeper) GetAllConfirmedSubmissions(ctx sdk.Context) []types.SubmissionKey {
	return k.getSubmissionsWithPrefix(ctx, types.ConfirmedIndexPrefix)
}

func (k Keeper) GetAllFinalizedSubmissions(ctx sdk.Context) []types.SubmissionKey {
	return k.getSubmissionsWithPrefix(ctx, types.FinalizedIndexPrefix)
}

// Callback to be called when btc light client tip change
func (k Keeper) OnTipChange(ctx sdk.Context) {
	k.checkUnconfirmed(ctx)
	k.checkConfirmed(ctx)
}
