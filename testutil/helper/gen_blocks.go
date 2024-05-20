package helper

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/merkle"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/babylonchain/babylon/testutil/datagen"
)

// adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.50.6/baseapp/abci_utils_test.go
func ExtendedCommitToLastCommit(ec abci.ExtendedCommitInfo) (abci.ExtendedCommitInfo, abci.CommitInfo, comet.BlockInfo) {
	// sort the extended commit info
	// below are copied from https://github.com/cosmos/cosmos-sdk/blob/v0.50.6/baseapp/abci_utils_test.go
	// Since v0.50.5 Cosmos SDK enforces certain order for vote extensions
	sort.SliceStable(ec.Votes, func(i, j int) bool {
		if ec.Votes[i].Validator.Power == ec.Votes[j].Validator.Power {
			return bytes.Compare(ec.Votes[i].Validator.Address, ec.Votes[j].Validator.Address) == -1
		}
		return ec.Votes[i].Validator.Power > ec.Votes[j].Validator.Power
	})

	// convert the extended commit info to last commit info
	lastCommit := abci.CommitInfo{
		Round: ec.Round,
		Votes: make([]abci.VoteInfo, len(ec.Votes)),
	}

	for i, vote := range ec.Votes {
		lastCommit.Votes[i] = abci.VoteInfo{
			Validator: abci.Validator{
				Address: vote.Validator.Address,
				Power:   vote.Validator.Power,
			},
		}
	}

	return ec, lastCommit, baseapp.NewBlockInfo(
		nil,
		nil,
		nil,
		lastCommit,
	)
}

func (h *Helper) genAndApplyEmptyBlock() error {
	prevHeight := h.App.LastBlockHeight()
	newHeight := prevHeight + 1

	// finalize block
	valSet, err := h.App.StakingKeeper.GetLastValidators(h.Ctx)
	if err != nil {
		return err
	}
	valhash := CalculateValHash(valSet)
	newHeader := cmttypes.Header{
		Height:             newHeight,
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
	}

	_, err = h.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             newHeader.Height,
		NextValidatorsHash: newHeader.NextValidatorsHash,
		Hash:               newHeader.Hash(),
	})
	if err != nil {
		return err
	}

	_, err = h.App.Commit()
	if err != nil {
		return err
	}

	if newHeight == 1 {
		// do it again
		// TODO: Figure out why when ctx height is 1, ApplyEmptyBlockWithVoteExtension
		// will still give ctx height 1 once, then start to increment
		return h.genAndApplyEmptyBlock()
	}

	return nil
}

func (h *Helper) ApplyEmptyBlockWithVoteExtension(r *rand.Rand) (sdk.Context, error) {
	emptyCtx := sdk.Context{}
	if h.App.LastBlockHeight() == 0 {
		if err := h.genAndApplyEmptyBlock(); err != nil {
			return emptyCtx, err
		}
	}
	valSetWithKeys := h.GenValidators
	prevHeight := h.App.LastBlockHeight()
	epoch := h.App.EpochingKeeper.GetEpoch(h.Ctx)
	newHeight := prevHeight + 1

	// 1. get previous vote extensions
	prevEpoch := epoch.EpochNumber
	blockHash := datagen.GenRandomBlockHash(r)
	extendedVotes, err := h.getExtendedVotesFromValSet(prevEpoch, uint64(prevHeight), blockHash, valSetWithKeys)
	if err != nil {
		return emptyCtx, err
	}
	extendedCommitInfo := abci.ExtendedCommitInfo{Votes: extendedVotes}
	_, lastCommitInfo, cometInfo := ExtendedCommitToLastCommit(extendedCommitInfo)

	// 2. create new header
	valSet, err := h.App.StakingKeeper.GetLastValidators(h.Ctx)
	if err != nil {
		return emptyCtx, err
	}
	valhash := CalculateValHash(valSet)
	newHeader := cmttypes.Header{
		Height:             newHeight,
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
		LastBlockID: cmttypes.BlockID{
			Hash: datagen.GenRandomByteArray(r, 32),
		},
	}
	h.Ctx = h.Ctx.WithHeaderInfo(header.Info{
		Height: newHeader.Height,
		Hash:   newHeader.Hash(),
	}).WithBlockHeader(*newHeader.ToProto()).WithCometInfo(cometInfo)

	// 3. prepare proposal with previous BLS sigs
	blockTxs := [][]byte{}
	ppRes, err := h.App.PrepareProposal(&abci.RequestPrepareProposal{
		LocalLastCommit: extendedCommitInfo,
		Height:          newHeight,
	})
	if err != nil {
		return emptyCtx, err
	}

	if len(ppRes.Txs) > 0 {
		blockTxs = ppRes.Txs
	}
	processRes, err := h.App.ProcessProposal(&abci.RequestProcessProposal{
		ProposedLastCommit: lastCommitInfo,
		Txs:                ppRes.Txs,
		Height:             newHeight,
	})
	if err != nil {
		return emptyCtx, err
	}
	if processRes.Status == abci.ResponseProcessProposal_REJECT {
		return emptyCtx, fmt.Errorf("rejected proposal")
	}

	// 4. finalize block
	resp, err := h.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Txs:                blockTxs,
		Height:             newHeader.Height,
		NextValidatorsHash: newHeader.NextValidatorsHash,
		Hash:               newHeader.Hash(),
	})
	if err != nil {
		return emptyCtx, err
	}

	newHeader.AppHash = resp.AppHash
	h.Ctx = h.Ctx.WithHeaderInfo(header.Info{
		Height:  newHeader.Height,
		AppHash: resp.AppHash,
	})

	_, err = h.App.Commit()
	if err != nil {
		return emptyCtx, err
	}

	return h.Ctx, nil
}

func (h *Helper) ApplyEmptyBlockWithValSet(r *rand.Rand, valSetWithKeys *datagen.GenesisValidators) (sdk.Context, error) {
	emptyCtx := sdk.Context{}
	if h.App.LastBlockHeight() == 0 {
		if err := h.genAndApplyEmptyBlock(); err != nil {
			return emptyCtx, err
		}
	}
	prevHeight := h.App.LastBlockHeight()
	epoch := h.App.EpochingKeeper.GetEpoch(h.Ctx)
	newHeight := prevHeight + 1

	// 1. get previous vote extensions
	prevEpoch := epoch.EpochNumber
	blockHash := datagen.GenRandomBlockHash(r)
	extendedVotes, err := h.getExtendedVotesFromValSet(prevEpoch, uint64(prevHeight), blockHash, valSetWithKeys)
	if err != nil {
		return emptyCtx, err
	}
	extendedCommitInfo := abci.ExtendedCommitInfo{Votes: extendedVotes}
	_, lastCommitInfo, cometInfo := ExtendedCommitToLastCommit(extendedCommitInfo)

	// 2. create new header
	valSet, err := h.App.StakingKeeper.GetLastValidators(h.Ctx)
	if err != nil {
		return emptyCtx, err
	}
	valhash := CalculateValHash(valSet)
	newHeader := cmttypes.Header{
		Height:             newHeight,
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
		LastBlockID: cmttypes.BlockID{
			Hash: datagen.GenRandomByteArray(r, 32),
		},
	}
	h.Ctx = h.Ctx.WithHeaderInfo(header.Info{
		Height: newHeader.Height,
		Hash:   newHeader.Hash(),
	}).WithBlockHeader(*newHeader.ToProto()).WithCometInfo(cometInfo)

	// 3. prepare proposal with previous BLS sigs
	blockTxs := [][]byte{}
	ppRes, err := h.App.PrepareProposal(&abci.RequestPrepareProposal{
		LocalLastCommit: abci.ExtendedCommitInfo{Votes: extendedVotes},
		Height:          newHeight,
	})
	if err != nil {
		return emptyCtx, err
	}

	if len(ppRes.Txs) > 0 {
		blockTxs = ppRes.Txs
	}
	processRes, err := h.App.ProcessProposal(&abci.RequestProcessProposal{
		ProposedLastCommit: lastCommitInfo,
		Txs:                ppRes.Txs,
		Height:             newHeight,
	})
	if err != nil {
		return emptyCtx, err
	}
	if processRes.Status == abci.ResponseProcessProposal_REJECT {
		return emptyCtx, fmt.Errorf("rejected proposal")
	}

	// 4. finalize block
	resp, err := h.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Txs:                blockTxs,
		Height:             newHeader.Height,
		NextValidatorsHash: newHeader.NextValidatorsHash,
		Hash:               newHeader.Hash(),
	})
	if err != nil {
		return emptyCtx, err
	}

	newHeader.AppHash = resp.AppHash
	h.Ctx = h.Ctx.WithHeaderInfo(header.Info{
		Height:  newHeader.Height,
		AppHash: resp.AppHash,
		Hash:    newHeader.Hash(),
	}).WithBlockHeader(*newHeader.ToProto())

	_, err = h.App.Commit()
	if err != nil {
		return emptyCtx, err
	}

	return h.Ctx, nil
}

func (h *Helper) ApplyEmptyBlockWithInvalidVoteExtensions(r *rand.Rand) (sdk.Context, error) {
	emptyCtx := sdk.Context{}
	if h.App.LastBlockHeight() == 0 {
		if err := h.genAndApplyEmptyBlock(); err != nil {
			return emptyCtx, err
		}
	}
	valSetWithKeys := h.GenValidators
	prevHeight := h.App.LastBlockHeight()
	epoch := h.App.EpochingKeeper.GetEpoch(h.Ctx)
	newHeight := prevHeight + 1

	// 1. get vote extensions with invalid BLS signature
	prevEpoch := epoch.EpochNumber
	blockHash := datagen.GenRandomBlockHash(r)
	extendedVotes, err := datagen.GenRandomVoteExtension(prevEpoch, uint64(prevHeight), blockHash, valSetWithKeys, r)
	if err != nil {
		return emptyCtx, err
	}
	extendedCommitInfo := abci.ExtendedCommitInfo{Votes: extendedVotes}
	_, lastCommitInfo, cometInfo := ExtendedCommitToLastCommit(extendedCommitInfo)

	res, err := h.App.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{
		Hash:          blockHash,
		Height:        prevHeight,
		VoteExtension: extendedVotes[0].VoteExtension,
	})
	if err != nil || !res.IsAccepted() {
		return emptyCtx, fmt.Errorf("invalid vote extension")
	}

	// 2. create new header
	valSet, err := h.App.StakingKeeper.GetLastValidators(h.Ctx)
	if err != nil {
		return emptyCtx, err
	}
	valhash := CalculateValHash(valSet)
	newHeader := cmttypes.Header{
		Height:             newHeight,
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
		LastBlockID: cmttypes.BlockID{
			Hash: datagen.GenRandomByteArray(r, 32),
		},
	}
	h.Ctx = h.Ctx.WithHeaderInfo(header.Info{
		Height: newHeader.Height,
		Hash:   newHeader.Hash(),
	}).WithBlockHeader(*newHeader.ToProto()).WithCometInfo(cometInfo)

	// 3. prepare proposal with previous BLS sigs
	blockTxs := [][]byte{}
	ppRes, err := h.App.PrepareProposal(&abci.RequestPrepareProposal{
		LocalLastCommit: extendedCommitInfo,
		Height:          newHeight,
	})
	if err != nil {
		return emptyCtx, err
	}

	if len(ppRes.Txs) > 0 {
		blockTxs = ppRes.Txs
	}
	processRes, err := h.App.ProcessProposal(&abci.RequestProcessProposal{
		ProposedLastCommit: lastCommitInfo,
		Txs:                ppRes.Txs,
		Height:             newHeight,
	})
	if err != nil {
		return emptyCtx, err
	}
	if processRes.Status == abci.ResponseProcessProposal_REJECT {
		return emptyCtx, fmt.Errorf("rejected proposal")
	}

	// 4. finalize block
	resp, err := h.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Txs:                blockTxs,
		Height:             newHeader.Height,
		NextValidatorsHash: newHeader.NextValidatorsHash,
		Hash:               newHeader.Hash(),
	})
	if err != nil {
		return emptyCtx, err
	}

	newHeader.AppHash = resp.AppHash
	h.Ctx = h.Ctx.WithHeaderInfo(header.Info{
		Height:  newHeader.Height,
		AppHash: resp.AppHash,
		Hash:    newHeader.Hash(),
	}).WithBlockHeader(*newHeader.ToProto())

	_, err = h.App.Commit()
	if err != nil {
		return emptyCtx, err
	}

	return h.Ctx, nil
}

func (h *Helper) ApplyEmptyBlockWithSomeInvalidVoteExtensions(r *rand.Rand) (sdk.Context, error) {
	emptyCtx := sdk.Context{}
	if h.App.LastBlockHeight() == 0 {
		if err := h.genAndApplyEmptyBlock(); err != nil {
			return emptyCtx, err
		}
	}
	valSetWithKeys := h.GenValidators
	prevHeight := h.App.LastBlockHeight()
	epoch := h.App.EpochingKeeper.GetEpoch(h.Ctx)
	newHeight := prevHeight + 1

	// 1. get previous vote extensions
	prevEpoch := epoch.EpochNumber
	blockHash := datagen.GenRandomBlockHash(r)
	extendedVotes, err := h.getExtendedVotesFromValSet(prevEpoch, uint64(prevHeight), blockHash, valSetWithKeys)
	if err != nil {
		return emptyCtx, err
	}
	extendedCommitInfo := abci.ExtendedCommitInfo{Votes: extendedVotes}
	_, lastCommitInfo, cometInfo := ExtendedCommitToLastCommit(extendedCommitInfo)

	// 2. create new header
	valSet, err := h.App.StakingKeeper.GetLastValidators(h.Ctx)
	if err != nil {
		return emptyCtx, err
	}
	valhash := CalculateValHash(valSet)
	newHeader := cmttypes.Header{
		Height:             newHeight,
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
		LastBlockID: cmttypes.BlockID{
			Hash: datagen.GenRandomByteArray(r, 32),
		},
	}
	h.Ctx = h.Ctx.WithHeaderInfo(header.Info{
		Height: newHeader.Height,
		Hash:   newHeader.Hash(),
	}).WithBlockHeader(*newHeader.ToProto()).WithCometInfo(cometInfo)

	// 3. prepare proposal with previous BLS sigs
	var blockTxs [][]byte
	if epoch.IsVoteExtensionProposal(h.Ctx) {
		// nullifies a subset of extended votes
		numEmptyVoteExts := len(extendedVotes)/3 - 1
		for i := 0; i < numEmptyVoteExts; i++ {
			extendedVotes[i].VoteExtension = datagen.GenRandomByteArray(r, uint64(r.Intn(10)))
		}
	}

	ppRes, err := h.App.PrepareProposal(&abci.RequestPrepareProposal{
		LocalLastCommit: extendedCommitInfo,
		Height:          newHeight,
	})
	if err != nil {
		return emptyCtx, err
	}
	blockTxs = ppRes.Txs

	processRes, err := h.App.ProcessProposal(&abci.RequestProcessProposal{
		ProposedLastCommit: lastCommitInfo,
		Txs:                blockTxs,
		Height:             newHeight,
	})
	if err != nil {
		return emptyCtx, err
	}
	if processRes.Status == abci.ResponseProcessProposal_REJECT {
		return emptyCtx, fmt.Errorf("rejected proposal")
	}

	// 4. finalize block
	resp, err := h.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Txs:                blockTxs,
		Height:             newHeader.Height,
		NextValidatorsHash: newHeader.NextValidatorsHash,
		Hash:               newHeader.Hash(),
	})
	if err != nil {
		return emptyCtx, err
	}

	newHeader.AppHash = resp.AppHash
	h.Ctx = h.Ctx.WithHeaderInfo(header.Info{
		Height:  newHeader.Height,
		AppHash: resp.AppHash,
		Hash:    newHeader.Hash(),
	}).WithBlockHeader(*newHeader.ToProto())

	_, err = h.App.Commit()
	if err != nil {
		return emptyCtx, err
	}

	return h.Ctx, nil
}

// CalculateValHash calculate validator hash and new header
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/simapp/test_helpers.go#L156-L163)
func CalculateValHash(valSet []stakingtypes.Validator) []byte {
	bzs := make([][]byte, len(valSet))
	for i, val := range valSet {
		consAddr, _ := val.GetConsAddr()
		bzs[i] = consAddr
	}
	return merkle.HashFromByteSlices(bzs)
}
