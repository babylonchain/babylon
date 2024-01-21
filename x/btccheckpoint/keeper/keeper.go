package keeper

import (
	"context"
	corestoretypes "cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/runtime"
	"math/big"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	txformat "github.com/babylonchain/babylon/btctxformatter"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc                  codec.BinaryCodec
		storeService         corestoretypes.KVStoreService
		tsKey                storetypes.StoreKey
		btcLightClientKeeper types.BTCLightClientKeeper
		checkpointingKeeper  types.CheckpointingKeeper
		incentiveKeeper      types.IncentiveKeeper
		powLimit             *big.Int
		authority            string
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
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,
	tsKey storetypes.StoreKey,
	bk types.BTCLightClientKeeper,
	ck types.CheckpointingKeeper,
	ik types.IncentiveKeeper,
	powLimit *big.Int,
	authority string,
) Keeper {

	return Keeper{
		cdc:                  cdc,
		storeService:         storeService,
		tsKey:                tsKey,
		btcLightClientKeeper: bk,
		checkpointingKeeper:  ck,
		incentiveKeeper:      ik,
		powLimit:             powLimit,
		authority:            authority,
	}
}

func (k Keeper) GetPowLimit() *big.Int {
	return k.powLimit
}

// GetExpectedTag retrerieves checkpoint tag from params and decodes it from
// hex string to bytes.
// NOTE: keeper could probably cache decoded tag, but it is rather improbable this function
// will ever be a bottleneck so it is not worth it.
func (k Keeper) GetExpectedTag(ctx context.Context) txformat.BabylonTag {
	tag := k.GetParams(ctx).CheckpointTag

	tagAsBytes, err := hex.DecodeString(tag)

	if err != nil {
		panic("Tag should always be valid")
	}

	return txformat.BabylonTag(tagAsBytes)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetBlockHeight(ctx context.Context, b *bbn.BTCHeaderHashBytes) (uint64, error) {
	return k.btcLightClientKeeper.BlockHeight(ctx, b)
}

func (k Keeper) headerDepth(ctx context.Context, headerHash *bbn.BTCHeaderHashBytes) (uint64, error) {
	blockDepth, err := k.btcLightClientKeeper.MainChainDepth(ctx, headerHash)

	if err != nil {
		// one of blocks is not known to light client
		return 0, submissionUnknownErr
	}
	return uint64(blockDepth), nil
}

// checkAncestors checks if there is at least one ancestor in previous epoch submissions
// previous epoch submission is considered ancestor when:
// - it is on main chain
// - its lowest depth is larger than highest depth of new submission
func (k Keeper) checkAncestors(
	ctx context.Context,
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

	if len(previousEpochData.Keys) == 0 {
		return types.ErrNoCheckpointsForPreviousEpoch
	}

	var haveDescendant = false

	for _, sk := range previousEpochData.Keys {
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

func (k Keeper) setBtcLightClientUpdated(ctx context.Context) {
	store := sdk.UnwrapSDKContext(ctx).TransientStore(k.tsKey)
	store.Set(types.GetBtcLightClientUpdatedKey(), []byte{1})
}

// BtcLightClientUpdated checks if btc light client was updated during block execution
func (k Keeper) BtcLightClientUpdated(ctx context.Context) bool {
	// transient store is cleared after each block execution, therfore if
	// BtcLightClientKey is set, it means setBtcLightClientUpdated was called during
	// current block execution
	store := sdk.UnwrapSDKContext(ctx).TransientStore(k.tsKey)
	lcUpdated := store.Get(types.GetBtcLightClientUpdatedKey())
	return len(lcUpdated) > 0
}

func (k Keeper) getLastFinalizedEpochNumber(ctx context.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	epoch, err := store.Get(types.GetLatestFinalizedEpochKey())
	if err != nil {
		panic(err)
	}
	if len(epoch) == 0 {
		return uint64(0)
	}

	return sdk.BigEndianToUint64(epoch)
}

func (k Keeper) setLastFinalizedEpochNumber(ctx context.Context, epoch uint64) {
	store := k.storeService.OpenKVStore(ctx)
	if err := store.Set(types.GetLatestFinalizedEpochKey(), sdk.Uint64ToBigEndian(epoch)); err != nil {
		panic(err)
	}
}

func (k Keeper) getEpochChanges(
	ctx context.Context,
	parentEpochBestSubmission *types.SubmissionBtcInfo,
	ed *types.EpochData) *epochChangesSummary {

	var submissionsToKeep []*types.SubmissionKey
	var submissionsToDelete []*types.SubmissionKey
	var currentEpochBestSubmission *types.SubmissionBtcInfo
	var bestSubmissionIdx int

	for i, sk := range ed.Keys {
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

// OnTipChange is the callback function to be called when btc light client tip changes
func (k Keeper) OnTipChange(ctx context.Context) {
	k.checkCheckpoints(ctx)
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

func (k Keeper) checkCheckpoints(ctx context.Context) {
	store := k.epochDataStore(ctx)

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

		if len(currentEpoch.Keys) == 0 {
			// current epoch does not have any submissions, so following one should also
			// not have any submissions, stop the processing.
			break
		}

		if currentEpoch.Status == types.Finalized {
			// current epoch is already finalized. This is our first epoch in iteration
			// just set parent info
			if len(currentEpoch.Keys) != 1 {
				panic("Finalized epoch must have only one valid submission")
			}

			subInfo, err := k.GetSubmissionBtcInfo(ctx, *currentEpoch.Keys[0])

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
		}

		if bestSubmissionStatus > currentEpoch.Status && currentEpoch.Status == types.Confirmed {
			// epoch just got finalized by best submission
			currentEpoch.Status = types.Finalized
			k.checkpointingKeeper.SetCheckpointFinalized(ctx, epoch)
			k.setLastFinalizedEpochNumber(ctx, epoch)
		}

		if currentEpoch.Status == types.Finalized {
			// trigger incentive module to distribute rewards to submitters/reporters
			k.rewardBTCTimestamping(ctx, epoch, &currentEpoch, epochChanges.BestSubmissionIdx)
			// delete all submissions except best one
			for i, sk := range currentEpoch.Keys {
				if i != epochChanges.BestSubmissionIdx {
					k.deleteSubmission(ctx, *sk)
				}
			}
			// leave only best submission key
			currentEpoch.Keys = []*types.SubmissionKey{&epochChanges.EpochBestSubmission.SubmissionKey}
		} else {
			// apply changes to epoch according to changes
			for _, sk := range epochChanges.SubmissionsToDelete {
				k.deleteSubmission(ctx, *sk)
			}
			currentEpoch.Keys = epochChanges.SubmissionsToKeep
		}

		parentEpochInfo = &epochInfo{bestSubmission: epochChanges.EpochBestSubmission}
		// save epoch with all applied changes
		store.Set(it.Key(), k.cdc.MustMarshal(&currentEpoch))
	}
}

func (k *Keeper) epochDataStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.EpochDataPrefix)
}
