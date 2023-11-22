package keeper

import (
	"bytes"
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/crypto/eots"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams updates the params
func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// AddFinalitySig adds a new vote to a given block
func (ms msgServer) AddFinalitySig(goCtx context.Context, req *types.MsgAddFinalitySig) (*types.MsgAddFinalitySigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// ensure the BTC validator exists
	btcVal, err := ms.BTCStakingKeeper.GetBTCValidator(ctx, req.ValBtcPk.MustMarshal())
	if err != nil {
		return nil, err
	}
	// ensure the BTC validator is not slashed at this time point
	// NOTE: it's possible that the BTC validator equivocates for height h, and the signature is processed at
	// height h' > h. In this case:
	// - Babylon should reject any new signature from this BTC validator, since it's known to be adversarial
	// - Babylon should set its voting power since height h'+1 to be zero, due to the same reason
	// - Babylon should NOT set its voting power between [h, h'] to be zero, since
	//   - Babylon BTC staking ensures safety upon 2f+1 votes, *even if* f of them are adversarial. This is
	//     because as long as a block gets 2f+1 votes, any other block with 2f+1 votes has a f+1 quorum
	//     intersection with this block, contradicting to the assumption and leading to the safety proof.
	//     This ensures slashable safety together with EOTS, thus does not undermine Babylon's security guarantee.
	//   - Due to this reason, when tallying a block, Babylon finalises this block upon 2f+1 votes. If we
	//     modify voting power table in the history, some finality decisions might be contradicting to the
	//     signature set and voting power table.
	//   - To fix the above issue, Babylon has to allow finalise and unfinalise blocks. However, this means
	//     Babylon will lose safety under an adaptive adversary corrupting even 1 validator. It can simply
	//     corrupt a new validator and equivocate a historical block over and over again, making a previous block
	//     unfinalisable forever
	if btcVal.IsSlashed() {
		return nil, bstypes.ErrBTCValAlreadySlashed
	}

	// ensure the BTC validator has voting power at this height
	valPK := req.ValBtcPk
	if ms.BTCStakingKeeper.GetVotingPower(ctx, valPK.MustMarshal(), req.BlockHeight) == 0 {
		return nil, types.ErrInvalidFinalitySig.Wrapf("the BTC validator %v does not have voting power at height %d", valPK.MustMarshal(), req.BlockHeight)
	}

	// ensure the BTC validator has not casted the same vote yet
	existingSig, err := ms.GetSig(ctx, req.BlockHeight, valPK)
	if err == nil && existingSig.Equals(req.FinalitySig) {
		return nil, types.ErrDuplicatedFinalitySig
	}

	// ensure the BTC validator has committed public randomness
	pubRand, err := ms.GetPubRand(ctx, valPK, req.BlockHeight)
	if err != nil {
		return nil, types.ErrPubRandNotFound
	}

	// verify EOTS signature w.r.t. public randomness
	valBTCPK, err := valPK.ToBTCPK()
	if err != nil {
		return nil, err
	}
	if err := eots.Verify(valBTCPK, pubRand.ToFieldVal(), req.MsgToSign(), req.FinalitySig.ToModNScalar()); err != nil {
		return nil, types.ErrInvalidFinalitySig.Wrapf("the EOTS signature is invalid: %v", err)
	}

	// verify whether the voted block is a fork or not
	indexedBlock, err := ms.GetBlock(ctx, req.BlockHeight)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(indexedBlock.AppHash, req.BlockAppHash) {
		// the BTC validator votes for a fork!

		// construct evidence
		evidence := &types.Evidence{
			ValBtcPk:                req.ValBtcPk,
			BlockHeight:             req.BlockHeight,
			PubRand:                 pubRand,
			CanonicalAppHash: indexedBlock.AppHash,
			CanonicalFinalitySig:    nil,
			ForkAppHash:      req.BlockAppHash,
			ForkFinalitySig:         req.FinalitySig,
		}

		// if this BTC validator has also signed canonical block, slash it
		canonicalSig, err := ms.GetSig(ctx, req.BlockHeight, valPK)
		if err == nil {
			//set canonial sig
			evidence.CanonicalFinalitySig = canonicalSig
			// slash this BTC validator, including setting its voting power to
			// zero, extracting its BTC SK, and emit an event
			ms.slashBTCValidator(ctx, req.ValBtcPk, evidence)
		}

		// save evidence
		ms.SetEvidence(ctx, evidence)

		// NOTE: we should NOT return error here, otherwise the state change triggered in this tx
		// (including the evidence) will be rolled back
		return &types.MsgAddFinalitySigResponse{}, nil
	}

	// this signature is good, add vote to DB
	ms.SetSig(ctx, req.BlockHeight, valPK, req.FinalitySig)

	// if this BTC validator has signed the canonical block before,
	// slash it via extracting its secret key, and emit an event
	if ms.HasEvidence(ctx, req.ValBtcPk, req.BlockHeight) {
		// the BTC validator has voted for a fork before!
		// If this evidence is at the same height as this signature, slash this BTC validator

		// get evidence
		evidence, err := ms.GetEvidence(ctx, req.ValBtcPk, req.BlockHeight)
		if err != nil {
			panic(fmt.Errorf("failed to get evidence despite HasEvidence returns true"))
		}

		// set canonical sig to this evidence
		evidence.CanonicalFinalitySig = req.FinalitySig
		ms.SetEvidence(ctx, evidence)

		// slash this BTC validator, including setting its voting power to
		// zero, extracting its BTC SK, and emit an event
		ms.slashBTCValidator(ctx, req.ValBtcPk, evidence)
	}

	return &types.MsgAddFinalitySigResponse{}, nil
}

// CommitPubRandList commits a list of EOTS public randomness
func (ms msgServer) CommitPubRandList(goCtx context.Context, req *types.MsgCommitPubRandList) (*types.MsgCommitPubRandListResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// ensure the request contains enough number of public randomness
	minPubRand := ms.GetParams(ctx).MinPubRand
	givenPubRand := len(req.PubRandList)
	if uint64(givenPubRand) < minPubRand {
		return nil, types.ErrTooFewPubRand.Wrapf("required minimum: %d, actual: %d", minPubRand, givenPubRand)
	}

	// ensure the BTC validator is registered
	valBTCPKBytes := req.ValBtcPk.MustMarshal()
	if !ms.BTCStakingKeeper.HasBTCValidator(ctx, valBTCPKBytes) {
		return nil, bstypes.ErrBTCValNotFound.Wrapf("the validator with BTC PK %v is not registered", valBTCPKBytes)
	}

	// this BTC validator has not commit any public randomness,
	// commit the given public randomness list and return
	if ms.IsFirstPubRand(ctx, req.ValBtcPk) {
		ms.SetPubRandList(ctx, req.ValBtcPk, req.StartHeight, req.PubRandList)
		return &types.MsgCommitPubRandListResponse{}, nil
	}

	// ensure height and req.StartHeight do not overlap, i.e., height < req.StartHeight
	height, _, err := ms.GetLastPubRand(ctx, req.ValBtcPk)
	if err != nil {
		return nil, err
	}
	if height >= req.StartHeight {
		return nil, types.ErrInvalidPubRand.Wrapf("the start height (%d) has overlap with the height of the highest public randomness (%d)", req.StartHeight, height)
	}

	// all good, commit the given public randomness list
	ms.SetPubRandList(ctx, req.ValBtcPk, req.StartHeight, req.PubRandList)
	return &types.MsgCommitPubRandListResponse{}, nil
}

// slashBTCValidator slashes a BTC validator with the given evidence
// including setting its voting power to zero, extracting its BTC SK,
// and emit an event
func (k Keeper) slashBTCValidator(ctx context.Context, valBtcPk *bbn.BIP340PubKey, evidence *types.Evidence) {
	// slash this BTC validator, i.e., set its voting power to zero
	if err := k.BTCStakingKeeper.SlashBTCValidator(ctx, valBtcPk.MustMarshal()); err != nil {
		panic(fmt.Errorf("failed to slash BTC validator: %v", err))
	}

	// emit slashing event
	eventSlashing := types.NewEventSlashedBTCValidator(evidence)
	if err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(eventSlashing); err != nil {
		panic(fmt.Errorf("failed to emit EventSlashedBTCValidator event: %w", err))
	}
}
