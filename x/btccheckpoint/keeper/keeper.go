package keeper

import (
	"fmt"
	"math"
	"math/big"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc                   codec.BinaryCodec
		storeKey              storetypes.StoreKey
		memKey                storetypes.StoreKey
		paramstore            paramtypes.Subspace
		btcLightClientKeeper  types.BTCLightClientKeeper
		checkpointingKeeper   types.CheckpointingKeeper
		powLimit              *big.Int
		expectedCheckpointTag txformat.BabylonTag
	}

	submissionBtcError string

	epochChangesSummary struct {
		SubmissionsToDelete []*types.SubmissionKey
		SubmissionsToKeep   []*types.SubmissionKey
		EpochBestSubmission *types.SubmissionBtcInfo
		BestSubmissionIdx   int
	}

	epochInfo struct {
		bestSubmission *types.SubmissionBtcInfo
	}
)

// Error interface for submissionBtcError
func (e submissionBtcError) Error() string {
	return string(e)
}

const (
	submissionUnknownErr submissionBtcError = submissionBtcError(
		"One of submission blocks is not known to btclightclient",
	)

	subbmisionOnForkErr submissionBtcError = submissionBtcError(
		"One of submission blocks is not on the btc mainchain ",
	)
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bk types.BTCLightClientKeeper,
	ck types.CheckpointingKeeper,
	powLimit *big.Int,
	expectedTag txformat.BabylonTag,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                   cdc,
		storeKey:              storeKey,
		memKey:                memKey,
		paramstore:            ps,
		btcLightClientKeeper:  bk,
		checkpointingKeeper:   ck,
		powLimit:              powLimit,
		expectedCheckpointTag: expectedTag,
	}
}

func (k Keeper) GetPowLimit() *big.Int {
	return k.powLimit
}

func (k Keeper) GetExpectedTag() txformat.BabylonTag {
	return k.expectedCheckpointTag
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetBlockHeight(ctx sdk.Context, b *bbn.BTCHeaderHashBytes) (uint64, error) {
	return k.btcLightClientKeeper.BlockHeight(ctx, b)
}

func (k Keeper) CheckHeaderIsOnMainChain(ctx sdk.Context, hash *bbn.BTCHeaderHashBytes) bool {
	depth, err := k.btcLightClientKeeper.MainChainDepth(ctx, hash)
	return err == nil && depth >= 0
}

func (k Keeper) headerDepth(ctx sdk.Context, headerHash *bbn.BTCHeaderHashBytes) (uint64, error) {
	blockDepth, err := k.btcLightClientKeeper.MainChainDepth(ctx, headerHash)

	if err != nil {
		// one of blocks is not known to light client
		return 0, submissionUnknownErr
	}

	if blockDepth < 0 {
		//  one of submission blocks is on fork, treat whole submission as being on fork
		return 0, subbmisionOnForkErr
	}

	return uint64(blockDepth), nil
}

func (k Keeper) checkSubmissionStatus(ctx sdk.Context, info *types.SubmissionBtcInfo) types.BtcStatus {
	subDepth := info.SubmissionDepth()
	if subDepth >= k.GetParams(ctx).CheckpointFinalizationTimeout {
		return types.Finalized
	} else if subDepth >= k.GetParams(ctx).BtcConfirmationDepth {
		return types.Confirmed
	} else {
		return types.Submitted
	}
}

func (k Keeper) GetSubmissionBtcInfo(ctx sdk.Context, sk types.SubmissionKey) (*types.SubmissionBtcInfo, error) {

	var lowest uint64 = math.MaxUint64
	var highest uint64 = uint64(0)
	var lowestIndestInMostFreshBlock uint32 = math.MaxUint32

	for _, tk := range sk.Key {
		d, err := k.headerDepth(ctx, tk.Hash)

		if err != nil {
			return nil, err
		}

		if d <= lowest {
			lowest = d
			if tk.Index < lowestIndestInMostFreshBlock {
				lowestIndestInMostFreshBlock = tk.Index
			}
		}

		if d > highest {
			highest = d
		}
	}

	return &types.SubmissionBtcInfo{
		SubmissionKey:      sk,
		OldestBlockDepth:   highest,
		YoungestBlockDepth: lowest,
		LatestTxIndex:      lowestIndestInMostFreshBlock,
	}, nil
}

func (k Keeper) GetCheckpointEpoch(ctx sdk.Context, c []byte) (uint64, error) {
	return k.checkpointingKeeper.CheckpointEpoch(ctx, c)
}

func (k Keeper) SubmissionExists(ctx sdk.Context, sk types.SubmissionKey) bool {
	return k.GetSubmissionData(ctx, sk) != nil
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

// checkAncestors checks if there is at least one ancestor in previous epoch submissions
// previous epoch submission is considered ancestor when:
// - it is on main chain
// - its lowest depth is larger than highest depth of new submission
func (k Keeper) checkAncestors(
	ctx sdk.Context,
	submisionEpoch uint64,
	newSubmissionInfo *types.SubmissionBtcInfo,
) error {

	if submisionEpoch <= 1 {
		// do not need to check ancestors for epoch 0 and 1
		return nil
	}

	// this is valid checkpoint for not initial epoch, we need to check previous epoch
	// checkpoints
	previousEpochData := k.GetEpochData(ctx, submisionEpoch-1)

	// First check if there are any checkpoints for previous epoch at all.
	if previousEpochData == nil {
		return types.ErrNoCheckpointsForPreviousEpoch
	}

	if len(previousEpochData.Key) == 0 {
		return types.ErrNoCheckpointsForPreviousEpoch
	}

	var haveDescendant = false

	for _, sk := range previousEpochData.Key {
		if len(sk.Key) < 2 {
			panic("Submission key composed of less than 2 transactions keys in database")
		}

		parentEpochSubmissionInfo, err := k.GetSubmissionBtcInfo(ctx, *sk)

		if err != nil {
			// Previous epoch submission block either landed on fork or was pruned
			// Submission will be pruned, so it should not be treated vaiable ancestor
			continue
		}

		if newSubmissionInfo.HappenedAfter(parentEpochSubmissionInfo) {
			// previous epoch submission most fresh block is deeper in the chain
			// than the new submission oldest block, therefore we can say there is
			// implicit parent-child relationship between submission blocks
			haveDescendant = true
			break
		}
	}

	if !haveDescendant {
		return types.ErrProvidedHeaderDoesNotHaveAncestor
	}

	return nil
}

func (k Keeper) saveEpochData(ctx sdk.Context, e uint64, ed *types.EpochData) {
	store := ctx.KVStore(k.storeKey)
	ek := types.GetEpochIndexKey(e)
	eb := k.cdc.MustMarshal(ed)
	store.Set(ek, eb)
}

// addEpochSubmission save given submission key and data to database and takes
// car of updating any necessary indexes.
// Provided submmission should be known to btclightclient and all of its blocks
// should be on btc main chaing as viewed by btclightclient
func (k Keeper) addEpochSubmission(
	ctx sdk.Context,
	epochNum uint64,
	sk types.SubmissionKey,
	sd types.SubmissionData,
	epochRawCheckpoint []byte,
) error {

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

	if ed.Status == types.Finalized {
		// we already finlized given epoch so we do not need any more submissions
		// TODO We should probably compare new submmission with the exisiting submission
		// which finalized the epoch. As it means we finalized epoch with not the best
		// submission possible
		return types.ErrEpochAlreadyFinalized
	}

	if len(ed.Key) == 0 {
		// it is first epoch submission inform checkpointing module about this fact
		k.checkpointingKeeper.SetCheckpointSubmitted(ctx, epochNum)
	}

	ed.AppendKey(sk)
	k.saveEpochData(ctx, epochNum, ed)
	k.saveSubmission(ctx, sk, sd)
	return nil
}

func (k Keeper) saveSubmission(ctx sdk.Context, sk types.SubmissionKey, sd types.SubmissionData) {
	store := ctx.KVStore(k.storeKey)
	kBytes := types.PrefixedSubmisionKey(k.cdc, &sk)
	sBytes := k.cdc.MustMarshal(&sd)
	store.Set(kBytes, sBytes)
}

func (k Keeper) deleteSubmission(ctx sdk.Context, sk types.SubmissionKey) {
	store := ctx.KVStore(k.storeKey)
	kBytes := types.PrefixedSubmisionKey(k.cdc, &sk)
	store.Delete(kBytes)
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

// Callback to be called when btc light client tip change
func (k Keeper) OnTipChange(ctx sdk.Context) {
	k.checkCheckpoints(ctx)
}

func (k Keeper) getLastFinalizedEpochNumber(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	epoch := store.Get(types.GetLatestFinalizedEpochKey())

	if len(epoch) == 0 {
		return uint64(0)
	}

	return sdk.BigEndianToUint64(epoch)
}

func (k Keeper) setLastFinalizedEpochNumber(ctx sdk.Context, epoch uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetLatestFinalizedEpochKey(), sdk.Uint64ToBigEndian(epoch))
}

func (k Keeper) getEpochChanges(
	ctx sdk.Context,
	parentEpochBestSubmission *types.SubmissionBtcInfo,
	ed *types.EpochData) *epochChangesSummary {

	var submissionsToKeep []*types.SubmissionKey
	var submissionsToDelete []*types.SubmissionKey
	var currentEpochBestSubmission *types.SubmissionBtcInfo
	var bestSubmissionIdx int

	for i, sk := range ed.Key {
		sk := sk
		if len(sk.Key) < 2 {
			panic("Submission key composed of less than 2 transactions keys in database")
		}

		submissionInfo, err := k.GetSubmissionBtcInfo(ctx, *sk)

		if err != nil {
			// submission no longer on main chain, mark it as to delete, and do not count
			// it as vaiable submission
			submissionsToDelete = append(submissionsToDelete, sk)
			continue
		}

		if parentEpochBestSubmission != nil && !submissionInfo.HappenedAfter(parentEpochBestSubmission) {
			// we have parent epoch info
			// make sure that this epoch submission deepest header is less deep that parent epoch
			// best submission depth
			submissionsToDelete = append(submissionsToDelete, sk)
			continue
		}

		// at this point submission is on main chain and after best submission from
		// previous epoch.Keep it
		submissionsToKeep = append(submissionsToKeep, sk)

		if currentEpochBestSubmission == nil {
			// we do not have info of best submission in this epoch. Set current submission
			// as best
			currentEpochBestSubmission = submissionInfo
			continue
		}

		if submissionInfo.IsBetterThan(currentEpochBestSubmission) {
			currentEpochBestSubmission = submissionInfo
			bestSubmissionIdx = i
		}
	}

	return &epochChangesSummary{
		SubmissionsToDelete: submissionsToDelete,
		SubmissionsToKeep:   submissionsToKeep,
		EpochBestSubmission: currentEpochBestSubmission,
		BestSubmissionIdx:   bestSubmissionIdx,
	}
}

func (k Keeper) clearEpochData(
	ctx sdk.Context,
	epoch []byte,
	epochDataStore prefix.Store,
	currentEpoch *types.EpochData) {

	for _, sk := range currentEpoch.Key {
		k.deleteSubmission(ctx, *sk)
	}
	currentEpoch.Key = []*types.SubmissionKey{}
	epochDataStore.Set(epoch, k.cdc.MustMarshal(currentEpoch))
}

// checkCheckpoints is the main function checking status of all submissions
// on btc chain as viewed through btc light client.
// Check works roughly as follows:
// 1. Iterate over epochs either from last finalized epoch or first epoch epoch
// ever, as for those epochs we do not need to check status of subbmissions of parent epoch.
//
// 2. For each epoch check status of every submissions:
//   - if the submission is still on the main chain
//   - if the deepest (best) submission of older epoch happened before given submission
//   - how deep is the submission
//
// 3 	Mark each submission which is not known to btc light client or is on btc light fork as to delete
//
// 4. In each epoch choose best submission, ie the one which is deepest on btc
//    chain. For depth determination, only the depth of the youngest block count
//    i.e if the submission is split between block with depth 1 and 2, then submission
//    depth = 1. In case of a draw i.e Two or more best submissions having the same
//    youngest block, tie is resolved by comparing tx index. Submission with lower
//    tx index is treated as better one.
//
// 5. After choosing best submission, the status of epoch is checked. If best
//    submission depth >= k deep, epoch is treated as confirmed. If depth >= w deep
//    epoch is treated as finalized.
//
// 6. If the epoch became finalized, delete all submissions except best one. If the
//		epoch is to finalized, delete all submissions which were marked as to delete
//
// 7. If the epoch loses all of its submissions, delete all submissions from child
// 		epoch as then we do not have parent for those.

func (k Keeper) checkCheckpoints(ctx sdk.Context) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.EpochDataPrefix,
	)

	lastFinalizedEpoch := k.getLastFinalizedEpochNumber(ctx)

	var startingEpoch []byte

	if lastFinalizedEpoch > 0 {
		// start iteration over epochs either from the first epoch or from the last
		// finalized epoch
		startingEpoch = sdk.Uint64ToBigEndian(lastFinalizedEpoch)
	}

	it := store.Iterator(startingEpoch, nil)
	defer it.Close()

	var parentEpochInfo *epochInfo

	for ; it.Valid(); it.Next() {
		var currentEpoch types.EpochData
		k.cdc.MustUnmarshal(it.Value(), &currentEpoch)
		epoch := sdk.BigEndianToUint64(it.Key())

		if len(currentEpoch.Key) == 0 {
			// current epoch does not have any submissions, so following one should also
			// not have any submissions, stop the processing.
			break
		}

		if currentEpoch.Status == types.Finalized {
			// current epoch is already finalized. This is our first epoch in iteration
			// just set parent info
			if len(currentEpoch.Key) != 1 {
				panic("Finalized epoch must have only one valid submission")
			}

			subInfo, err := k.GetSubmissionBtcInfo(ctx, *currentEpoch.Key[0])

			if err != nil {
				panic("Finalized epoch submission must be on main chain")
			}

			parentEpochInfo = &epochInfo{
				bestSubmission: subInfo,
			}

			continue
		}
		// At this point we known that current epoch has at least 1 submission, and is
		// either Submitted or Confirmed

		if parentEpochInfo != nil && parentEpochInfo.bestSubmission == nil {
			// a bit of special case, when parent epoch lost all its submissions but current
			// epoch was already in submitted/confirmed states.
			// We need to delete all submissions, inform checkpointing module that this happened
			// and set epoch to set epoch to signed

			k.clearEpochData(ctx, it.Key(), store, &currentEpoch)
			k.checkpointingKeeper.SetCheckpointForgotten(ctx, epoch)
			// set parent epoch with empty best submission, so child epoch will also
			// get clearead
			parentEpochInfo = &epochInfo{}
			continue
		}

		var epochChanges *epochChangesSummary
		if parentEpochInfo == nil {
			// do not have parent epoch info, so this is first epoch, and we do not need
			// to validate ancestry
			epochChanges = k.getEpochChanges(ctx, nil, &currentEpoch)
		} else {
			// parentEpochInfo.bestSubmission will never be nil at this point
			epochChanges = k.getEpochChanges(ctx, parentEpochInfo.bestSubmission, &currentEpoch)
		}

		if len(epochChanges.SubmissionsToKeep) == 0 {
			// epoch lost all submissions clear it and inform checkpointing about it
			k.clearEpochData(ctx, it.Key(), store, &currentEpoch)
			k.checkpointingKeeper.SetCheckpointForgotten(ctx, epoch)
			// set parent epoch with empty best submission, so child epoch will also
			// get clearead
			parentEpochInfo = &epochInfo{}
			continue
		}

		// there is at least one submission in the epoch, check its current btc status
		bestSubmissionStatus := k.checkSubmissionStatus(ctx, epochChanges.EpochBestSubmission)

		if bestSubmissionStatus > currentEpoch.Status && currentEpoch.Status == types.Submitted {
			// epoch just got confirmed by best submission
			currentEpoch.Status = types.Confirmed
			k.checkpointingKeeper.SetCheckpointConfirmed(ctx, epoch)
			// TODO This is the place to check other submissions and pay up rewards.
		}

		if bestSubmissionStatus > currentEpoch.Status && currentEpoch.Status == types.Confirmed {
			// epoch just got finalized by best submission
			currentEpoch.Status = types.Finalized
			k.checkpointingKeeper.SetCheckpointFinalized(ctx, epoch)
			k.setLastFinalizedEpochNumber(ctx, epoch)
		}

		if currentEpoch.Status == types.Finalized {
			for i, sk := range currentEpoch.Key {
				// delete all submissions except best one
				if i != epochChanges.BestSubmissionIdx {
					k.deleteSubmission(ctx, *sk)
				}
				currentEpoch.Key = []*types.SubmissionKey{&epochChanges.EpochBestSubmission.SubmissionKey}
			}
		} else {
			// applay changes to epoch according to changes
			for _, sk := range epochChanges.SubmissionsToDelete {
				k.deleteSubmission(ctx, *sk)
			}
			currentEpoch.Key = epochChanges.SubmissionsToKeep
		}

		parentEpochInfo = &epochInfo{bestSubmission: epochChanges.EpochBestSubmission}
		// save epoch with all applied changes
		store.Set(it.Key(), k.cdc.MustMarshal(&currentEpoch))
	}
}
