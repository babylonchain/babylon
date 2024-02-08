package checkpointing

import (
	"fmt"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/checkpointing/keeper"
	ckpttypes "github.com/babylonchain/babylon/x/checkpointing/types"
)

// VoteExtensionHandler defines a BLS-based vote extension handlers for Babylon.
type VoteExtensionHandler struct {
	logger     log.Logger
	ckptKeeper *keeper.Keeper
	valStore   baseapp.ValidatorStore
}

func NewVoteExtensionHandler(logger log.Logger, ckptKeeper *keeper.Keeper) *VoteExtensionHandler {
	return &VoteExtensionHandler{logger: logger, ckptKeeper: ckptKeeper, valStore: ckptKeeper}
}

func (h *VoteExtensionHandler) SetHandlers(bApp *baseapp.BaseApp) {
	bApp.SetExtendVoteHandler(h.ExtendVote())
	bApp.SetVerifyVoteExtensionHandler(h.VerifyVoteExtension())
}

// ExtendVote sends a BLS signature as a vote extension
// the signature is signed over the hash of the last
// block of the current epoch
func (h *VoteExtensionHandler) ExtendVote() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		k := h.ckptKeeper
		// the returned response MUST not be nil
		emptyRes := &abci.ResponseExtendVote{VoteExtension: []byte{}}

		epoch := k.GetEpoch(ctx)
		// BLS vote extension is only applied at the last block of the current epoch
		if !epoch.IsLastBlockByHeight(req.Height) {
			return emptyRes, nil
		}

		// 1. check if itself is the validator as the BLS sig is only signed
		// when the node itself is a validator
		signer := k.GetBLSSignerAddress()
		curValSet := k.GetValidatorSet(ctx, epoch.EpochNumber)
		_, _, err := curValSet.FindValidatorWithIndex(signer)
		if err != nil {
			// not being a validator is not an error
			h.logger.Info("the BLS signer is not in the validator set", "signer_address", signer.String())
			return emptyRes, nil
		}

		// 2. sign BLS signature
		blsSig, err := k.SignBLS(epoch.EpochNumber, req.Hash)
		if err != nil {
			// the returned error will lead to panic
			return emptyRes, fmt.Errorf("failed to sign BLS signature at epoch %v, height %v",
				epoch.EpochNumber, req.Height)
		}

		// 3. build vote extension
		ve := &ckpttypes.VoteExtension{
			Signer:           signer.String(),
			ValidatorAddress: k.GetValidatorAddress().String(),
			BlockHash:        req.Hash,
			EpochNum:         epoch.EpochNumber,
			Height:           uint64(req.Height),
			BlsSig:           &blsSig,
		}
		bz, err := ve.Marshal()
		if err != nil {
			return emptyRes, fmt.Errorf("failed to encode vote extension: %w", err)
		}

		h.logger.Info("successfully sent BLS signature in vote extension",
			"epoch", epoch.EpochNumber, "height", req.Height)

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

// VerifyVoteExtension verifies the BLS sig within the vote extension
func (h *VoteExtensionHandler) VerifyVoteExtension() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		k := h.ckptKeeper
		resAccept := &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}
		resReject := &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}

		epoch := k.GetEpoch(ctx)
		// BLS vote extension is only applied at the last block of the current epoch
		if !epoch.IsLastBlockByHeight(req.Height) {
			return resAccept, nil
		}

		if len(req.VoteExtension) == 0 {
			h.logger.Error("received empty vote extension", "height", req.Height)
			return resReject, nil
		}

		var ve ckpttypes.VoteExtension
		if err := ve.Unmarshal(req.VoteExtension); err != nil {
			h.logger.Error("failed to unmarshal vote extension", "err", err, "height", req.Height)
			return resReject, nil
		}

		// 1. verify epoch number
		if epoch.EpochNumber != ve.EpochNum {
			h.logger.Error("invalid epoch number in the vote extension",
				"want", epoch.EpochNumber, "got", ve.EpochNum)
			return resReject, nil
		}

		// 2. ensure the validator address in the BLS sig matches the signer of the vote extension
		// this prevents validators use valid BLS from another validator
		blsSig := ve.ToBLSSig()
		extensionSigner := sdk.ValAddress(req.ValidatorAddress).String()
		if extensionSigner != blsSig.ValidatorAddress {
			h.logger.Error("the vote extension signer does not match the BLS signature signer",
				"extension signer", extensionSigner, "BLS signer", blsSig.SignerAddress)
			return resReject, nil
		}

		// 3. verify signing hash
		if !blsSig.BlockHash.Equal(req.Hash) {
			// processed BlsSig message is for invalid last commit hash
			h.logger.Error("in valid block ID in BLS sig", "want", req.Hash, "got", blsSig.BlockHash)
			return resReject, nil
		}

		// 4. verify the validity of the BLS signature
		if err := k.VerifyBLSSig(ctx, blsSig); err != nil {
			// Note: reject this vote extension as BLS is invalid
			// this will also reject the corresponding PreCommit vote
			h.logger.Error("invalid BLS sig in vote extension",
				"err", err,
				"height", req.Height,
				"epoch", epoch.EpochNumber,
			)
			return resReject, nil
		}

		h.logger.Info("successfully verified vote extension",
			"signer_address", ve.Signer,
			"height", req.Height,
			"epoch", epoch.EpochNumber,
		)

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}
