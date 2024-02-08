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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if err := req.Params.Validate(); err != nil {
		return nil, govtypes.ErrInvalidProposalMsg.Wrapf("invalid parameter: %v", err)
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

	// ensure the finality provider exists
	fp, err := ms.BTCStakingKeeper.GetFinalityProvider(ctx, req.FpBtcPk)
	if err != nil {
		return nil, err
	}
	// ensure the finality provider is not slashed at this time point
	// NOTE: it's possible that the finality provider equivocates for height h, and the signature is processed at
	// height h' > h. In this case:
	// - Babylon should reject any new signature from this finality provider, since it's known to be adversarial
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
	//     Babylon will lose safety under an adaptive adversary corrupting even 1 finality provider. It can simply
	//     corrupt a new finality provider and equivocate a historical block over and over again, making a previous block
	//     unfinalisable forever
	if fp.IsSlashed() {
		return nil, bstypes.ErrFpAlreadySlashed
	}

	// ensure the finality provider has voting power at this height
	if req.FpBtcPk == nil {
		return nil, types.ErrInvalidFinalitySig.Wrap("empty finality provider BTC PK")
	}
	fpPK, err := bbn.NewBIP340PubKey(req.FpBtcPk)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if ms.BTCStakingKeeper.GetVotingPower(ctx, *fpPK, req.BlockHeight) == 0 {
		return nil, types.ErrInvalidFinalitySig.Wrapf("the finality provider %v does not have voting power at height %d", fpPK, req.BlockHeight)
	}

	// ensure the finality provider has not cast the same vote yet
	if req.FinalitySig == nil {
		return nil, types.ErrInvalidFinalitySig.Wrap("empty finality signature")
	}
	existingSig, err := ms.GetSig(ctx, req.BlockHeight, fpPK)
	if err == nil && existingSig.Equals(req.FinalitySig) {
		ms.Logger(ctx).Debug("Received duplicated finiality vote", "block height", req.BlockHeight, "finality provider", req.FpBtcPk)
		// exactly same vote alreay exists, return success to the provider
		return &types.MsgAddFinalitySigResponse{}, nil
	}

	// ensure the finality provider has committed public randomness
	pubRand, err := ms.GetPubRand(ctx, fpPK, req.BlockHeight)
	if err != nil {
		return nil, types.ErrPubRandNotFound
	}

	// verify EOTS signature w.r.t. public randomness
	fpBTCPK, err := fpPK.ToBTCPK()
	if err != nil {
		return nil, err
	}
	if err := eots.Verify(fpBTCPK, pubRand.ToFieldVal(), req.MsgToSign(), req.FinalitySig.ToModNScalar()); err != nil {
		return nil, types.ErrInvalidFinalitySig.Wrapf("the EOTS signature is invalid: %v", err)
	}

	// verify whether the voted block is a fork or not
	indexedBlock, err := ms.GetBlock(ctx, req.BlockHeight)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(indexedBlock.AppHash, req.BlockAppHash) {
		// the finality provider votes for a fork!

		// construct evidence
		evidence := &types.Evidence{
			FpBtcPk:              fpPK,
			BlockHeight:          req.BlockHeight,
			PubRand:              pubRand,
			CanonicalAppHash:     indexedBlock.AppHash,
			CanonicalFinalitySig: nil,
			ForkAppHash:          req.BlockAppHash,
			ForkFinalitySig:      req.FinalitySig,
		}

		// if this finality provider has also signed canonical block, slash it
		canonicalSig, err := ms.GetSig(ctx, req.BlockHeight, fpPK)
		if err == nil {
			//set canonial sig
			evidence.CanonicalFinalitySig = canonicalSig
			// slash this finality provider, including setting its voting power to
			// zero, extracting its BTC SK, and emit an event
			ms.slashFinalityProvider(ctx, fpPK, evidence)
		}

		// save evidence
		ms.SetEvidence(ctx, evidence)

		// NOTE: we should NOT return error here, otherwise the state change triggered in this tx
		// (including the evidence) will be rolled back
		return &types.MsgAddFinalitySigResponse{}, nil
	}

	// this signature is good, add vote to DB
	ms.SetSig(ctx, req.BlockHeight, fpPK, req.FinalitySig)

	// if this finality provider has signed the canonical block before,
	// slash it via extracting its secret key, and emit an event
	if ms.HasEvidence(ctx, fpPK, req.BlockHeight) {
		// the finality provider has voted for a fork before!
		// If this evidence is at the same height as this signature, slash this finality provider

		// get evidence
		evidence, err := ms.GetEvidence(ctx, fpPK, req.BlockHeight)
		if err != nil {
			panic(fmt.Errorf("failed to get evidence despite HasEvidence returns true"))
		}

		// set canonical sig to this evidence
		evidence.CanonicalFinalitySig = req.FinalitySig
		ms.SetEvidence(ctx, evidence)

		// slash this finality provider, including setting its voting power to
		// zero, extracting its BTC SK, and emit an event
		ms.slashFinalityProvider(ctx, fpPK, evidence)
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

	// ensure the finality provider is registered
	fpPK, err := bbn.NewBIP340PubKey(req.FpBtcPk)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	fpBTCPKBytes := req.FpBtcPk
	if !ms.BTCStakingKeeper.HasFinalityProvider(ctx, fpBTCPKBytes) {
		return nil, bstypes.ErrFpNotFound.Wrapf("the finality provider with BTC PK %v is not registered", fpBTCPKBytes)
	}

	// this finality provider has not commit any public randomness,
	// commit the given public randomness list and return
	if ms.IsFirstPubRand(ctx, fpPK) {
		ms.SetPubRandList(ctx, fpPK, req.StartHeight, req.PubRandList)
		return &types.MsgCommitPubRandListResponse{}, nil
	}

	// ensure height and req.StartHeight do not overlap, i.e., height < req.StartHeight
	height, _, err := ms.GetLastPubRand(ctx, fpPK)
	if err != nil {
		return nil, err
	}
	if height >= req.StartHeight {
		return nil, types.ErrInvalidPubRand.Wrapf("the start height (%d) has overlap with the height of the highest public randomness (%d)", req.StartHeight, height)
	}

	// verify signature over the list
	if err := req.VerifySig(); err != nil {
		return nil, types.ErrInvalidPubRand.Wrapf("invalid signature over the public randomness list: %v", err)
	}

	// all good, commit the given public randomness list
	ms.SetPubRandList(ctx, fpPK, req.StartHeight, req.PubRandList)
	return &types.MsgCommitPubRandListResponse{}, nil
}

// slashFinalityProvider slashes a finality provider with the given evidence
// including setting its voting power to zero, extracting its BTC SK,
// and emit an event
func (k Keeper) slashFinalityProvider(ctx context.Context, fpBtcPk *bbn.BIP340PubKey, evidence *types.Evidence) {
	// slash this finality provider, i.e., set its voting power to zero
	if err := k.BTCStakingKeeper.SlashFinalityProvider(ctx, fpBtcPk.MustMarshal()); err != nil {
		panic(fmt.Errorf("failed to slash finality provider: %v", err))
	}

	// emit slashing event
	eventSlashing := types.NewEventSlashedFinalityProvider(evidence)
	if err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(eventSlashing); err != nil {
		panic(fmt.Errorf("failed to emit EventSlashedFinalityProvider event: %w", err))
	}
}
