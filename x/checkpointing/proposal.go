package checkpointing

import (
	"fmt"
	"slices"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/babylonchain/babylon/x/checkpointing/keeper"
	ckpttypes "github.com/babylonchain/babylon/x/checkpointing/types"
)

const defaultInjectedTxIndex = 0

type ProposalHandler struct {
	logger                        log.Logger
	ckptKeeper                    *keeper.Keeper
	valStore                      baseapp.ValidatorStore
	txVerifier                    baseapp.ProposalTxVerifier
	defaultPrepareProposalHandler sdk.PrepareProposalHandler
	defaultProcessProposalHandler sdk.ProcessProposalHandler
}

func NewProposalHandler(logger log.Logger, ckptKeeper *keeper.Keeper, mp mempool.Mempool, txVerifier baseapp.ProposalTxVerifier) *ProposalHandler {
	defaultHandler := baseapp.NewDefaultProposalHandler(mp, txVerifier)
	return &ProposalHandler{
		logger:                        logger,
		ckptKeeper:                    ckptKeeper,
		valStore:                      ckptKeeper,
		txVerifier:                    txVerifier,
		defaultPrepareProposalHandler: defaultHandler.PrepareProposalHandler(),
		defaultProcessProposalHandler: defaultHandler.ProcessProposalHandler(),
	}
}

func (h *ProposalHandler) SetHandlers(bApp *baseapp.BaseApp) {
	bApp.SetPrepareProposal(h.PrepareProposal())
	bApp.SetProcessProposal(h.ProcessProposal())
	bApp.SetPreBlocker(h.PreBlocker())
}

// PrepareProposal examines the vote extensions from the previous block, accumulates
// them into a checkpoint, and injects the checkpoint into the current proposal
// as a special tx
// Warning: the returned error of the handler will cause panic of the proposer,
// therefore we only return error when something really wrong happened
func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		// call default handler first to do basic validation
		res, err := h.defaultPrepareProposalHandler(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed in default PrepareProposal handler: %w", err)
		}

		k := h.ckptKeeper
		proposalTxs := res.Txs
		proposalRes := &abci.ResponsePrepareProposal{Txs: proposalTxs}

		epoch := k.GetEpoch(ctx)
		// BLS signatures are sent in the last block of the previous epoch,
		// so they should be aggregated in the first block of the new epoch
		// and no BLS signatures are send in epoch 0
		if !epoch.IsVoteExtensionProposal(ctx) {
			return proposalRes, nil
		}

		if len(req.LocalLastCommit.Votes) == 0 {
			return proposalRes, fmt.Errorf("no extended votes received from the last block")
		}

		// 1. verify the validity of vote extensions (2/3 majority is achieved)
		err = baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit)
		if err != nil {
			return proposalRes, fmt.Errorf("invalid vote extensions: %w", err)
		}

		// 2. build a checkpoint for the previous epoch
		// Note: the epoch has not increased yet, so
		// we can use the current epoch
		ckpt, err := h.buildCheckpointFromVoteExtensions(ctx, epoch.EpochNumber, req.LocalLastCommit.Votes)
		if err != nil {
			return proposalRes, fmt.Errorf("failed to build checkpoint from vote extensions: %w", err)
		}

		// 3. inject a "fake" tx into the proposal s.t. validators can decode, verify the checkpoint
		injectedCkpt := &ckpttypes.InjectedCheckpoint{
			Ckpt:               ckpt,
			ExtendedCommitInfo: &req.LocalLastCommit,
		}
		injectedVoteExtTx, err := injectedCkpt.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to encode vote extensions into a special tx: %w", err)
		}
		proposalTxs = slices.Insert(proposalTxs, defaultInjectedTxIndex, [][]byte{injectedVoteExtTx}...)

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

func (h *ProposalHandler) buildCheckpointFromVoteExtensions(ctx sdk.Context, epoch uint64, extendedVotes []abci.ExtendedVoteInfo) (*ckpttypes.RawCheckpointWithMeta, error) {
	prevBlockID, err := h.findLastBlockHash(extendedVotes)
	if err != nil {
		return nil, err
	}
	ckpt := ckpttypes.NewCheckpointWithMeta(ckpttypes.NewCheckpoint(epoch, prevBlockID), ckpttypes.Accumulating)
	validBLSSigs := h.getValidBlsSigs(ctx, extendedVotes)
	vals := h.ckptKeeper.GetValidatorSet(ctx, epoch)
	totalPower := h.ckptKeeper.GetTotalVotingPower(ctx, epoch)
	// TODO: maybe we don't need to verify BLS sigs anymore as they are already
	//  verified by VerifyVoteExtension
	for _, sig := range validBLSSigs {
		signerAddress, err := sdk.ValAddressFromBech32(sig.SignerAddress)
		if err != nil {
			h.logger.Error(
				"skip invalid BLS sig",
				"invalid signer address", sig.SignerAddress,
				"err", err,
			)
			continue
		}
		signerBlsKey, err := h.ckptKeeper.GetBlsPubKey(ctx, signerAddress)
		if err != nil {
			h.logger.Error(
				"skip invalid BLS sig",
				"can't find BLS public key", err,
			)
			continue
		}
		err = ckpt.Accumulate(vals, signerAddress, signerBlsKey, *sig.BlsSig, totalPower)
		if err != nil {
			h.logger.Error(
				"skip invalid BLS sig",
				"accumulation failed", err,
			)
			continue
		}
		// sufficient voting power is accumulated
		if ckpt.Status == ckpttypes.Sealed {
			break
		}
	}
	if ckpt.Status != ckpttypes.Sealed {
		return nil, fmt.Errorf("insufficient voting power to build the checkpoint")
	}

	return ckpt, nil
}

func (h *ProposalHandler) getValidBlsSigs(ctx sdk.Context, extendedVotes []abci.ExtendedVoteInfo) []ckpttypes.BlsSig {
	k := h.ckptKeeper
	validBLSSigs := make([]ckpttypes.BlsSig, 0, len(extendedVotes))
	for _, voteInfo := range extendedVotes {
		veBytes := voteInfo.VoteExtension
		if len(veBytes) == 0 {
			h.logger.Error("received empty vote extension", "validator", voteInfo.Validator.String())
			continue
		}
		var ve ckpttypes.VoteExtension
		if err := ve.Unmarshal(veBytes); err != nil {
			h.logger.Error("failed to unmarshal vote extension", "err", err)
			continue
		}
		sig := ve.ToBLSSig()

		if err := k.VerifyBLSSig(ctx, sig); err != nil {
			h.logger.Error("invalid BLS signature", "err", err)
			continue
		}

		validBLSSigs = append(validBLSSigs, *sig)
	}

	return validBLSSigs
}

// findLastBlockHash finds the last block hash from the first vote extension
// this is a workaround that the last block hash can't be obtained from the context
// this is also safe as the BLS sig is already verified by VerifyVoteExtension
func (h *ProposalHandler) findLastBlockHash(extendedVotes []abci.ExtendedVoteInfo) ([]byte, error) {
	var ve ckpttypes.VoteExtension
	if err := ve.Unmarshal(extendedVotes[0].VoteExtension); err != nil {
		return nil, err
	}
	return ve.ToBLSSig().BlockHash.MustMarshal(), nil
}

// ProcessProposal examines the checkpoint in the injected tx of the proposal
// Warning: the returned error of the handler will cause panic of the node,
// therefore we only return error when something really wrong happened
func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		resAccept := &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}
		resReject := &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}

		k := h.ckptKeeper

		epoch := k.GetEpoch(ctx)
		// BLS signatures are sent in the last block of the previous epoch,
		// so they should be aggregated in the first block of the new epoch
		// and no BLS signatures are send in epoch 0
		if epoch.IsVoteExtensionProposal(ctx) {
			// 1. extract the special tx containing the checkpoint
			injectedCkpt, err := extractInjectedCheckpoint(req.Txs)
			if err != nil {
				h.logger.Error("cannot get injected checkpoint", "err", err)
				// should not return error here as error will cause panic
				return resReject, nil
			}

			for i, v := range injectedCkpt.ExtendedCommitInfo.Votes {
				var ve ckpttypes.VoteExtension
				if err := ve.Unmarshal(v.VoteExtension); err != nil {
					return nil, err
				}
				if len(ve.BlockHash) == 0 {
					h.logger.Error("found empty block hash in the vote extension",
						"height", req.Height, "epoch", epoch.EpochNumber,
						"index", i)
				}
			}

			// 2. remove the special tx from the request so that
			// the rest of the txs can be handled by the default handler
			req.Txs, err = removeInjectedTx(req.Txs)
			if err != nil {
				// should not return error here as error will cause panic
				h.logger.Error("failed to remove injected tx from request: %w", err)
				return resReject, nil
			}

			// 3. verify the validity of the vote extension (2/3 majority is achieved)
			err = baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), *injectedCkpt.ExtendedCommitInfo)
			if err != nil {
				// the returned err will lead to panic as something very wrong happened during consensus
				return resReject, err
			}

			// 4. rebuild the checkpoint from vote extensions and compare it with
			// the injected checkpoint
			// Note: this is needed because LastBlockID is not available here so that
			// we can't verify whether the injected checkpoint is signing the correct
			// LastBlockID
			ckpt, err := h.buildCheckpointFromVoteExtensions(ctx, epoch.EpochNumber, injectedCkpt.ExtendedCommitInfo.Votes)
			if err != nil {
				// should not return error here as error will cause panic
				h.logger.Error("invalid vote extensions: %w", err)
				return resReject, nil
			}
			// TODO it is possible that although the checkpoints do not match but the injected
			//  checkpoint is still valid. This indicates the existence of a fork (>1/3 malicious voting power)
			//  and we should probably send an alarm and stall the blockchain
			if !ckpt.Equal(injectedCkpt.Ckpt) {
				// should not return error here as error will cause panic
				h.logger.Error("invalid checkpoint in vote extension tx", "err", err)
				return resReject, nil
			}
		}

		// 5. verify the rest of the txs using the default handler
		res, err := h.defaultProcessProposalHandler(ctx, req)
		if err != nil {
			return resReject, fmt.Errorf("failed in default ProcessProposal handler: %w", err)
		}
		if !res.IsAccepted() {
			h.logger.Error("the proposal is rejected by default ProcessProposal handler",
				"height", req.Height, "epoch", epoch.EpochNumber)
			return resReject, nil
		}

		return resAccept, nil
	}
}

// PreBlocker extracts the checkpoint from the injected tx and stores it in
// the application
// no more validation is needed as it is already done in ProcessProposal
func (h *ProposalHandler) PreBlocker() sdk.PreBlocker {
	return func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		k := h.ckptKeeper
		res := &sdk.ResponsePreBlock{}

		epoch := k.GetEpoch(ctx)
		// BLS signatures are sent in the last block of the previous epoch,
		// so they should be aggregated in the first block of the new epoch
		// and no BLS signatures are send in epoch 0
		if !epoch.IsVoteExtensionProposal(ctx) {
			return res, nil
		}

		// 1. extract the special tx containing BLS sigs
		injectedCkpt, err := extractInjectedCheckpoint(req.Txs)
		if err != nil {
			return res, fmt.Errorf("failed to get extract injected checkpoint from the tx set: %w", err)
		}

		// 2. update checkpoint
		if err := k.SealCheckpoint(ctx, injectedCkpt.Ckpt); err != nil {
			return res, fmt.Errorf("failed to update checkpoint: %w", err)
		}

		return res, nil
	}
}

// extractInjectedCheckpoint extracts the injected checkpoint from the tx set
func extractInjectedCheckpoint(txs [][]byte) (*ckpttypes.InjectedCheckpoint, error) {
	if len(txs) < defaultInjectedTxIndex+1 {
		return nil, fmt.Errorf("the tx set does not contain the injected tx")
	}

	injectedTx := txs[defaultInjectedTxIndex]

	if len(injectedTx) == 0 {
		return nil, fmt.Errorf("err in PreBlocker: the injected vote extensions tx is empty")
	}

	var injectedCkpt ckpttypes.InjectedCheckpoint
	if err := injectedCkpt.Unmarshal(injectedTx); err != nil {
		return nil, fmt.Errorf("failed to decode injected vote extension tx: %w", err)
	}

	return &injectedCkpt, nil
}

// removeInjectedTx removes the injected tx from the tx set
func removeInjectedTx(txs [][]byte) ([][]byte, error) {
	if len(txs) < defaultInjectedTxIndex+1 {
		return nil, fmt.Errorf("the tx set does not contain the injected tx")
	}

	txs = append(txs[:defaultInjectedTxIndex], txs[defaultInjectedTxIndex+1:]...)

	return txs, nil
}
