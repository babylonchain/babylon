package keeper

import (
	"bytes"
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/crypto/eots"
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
	if !bytes.Equal(indexedBlock.LastCommitHash, req.BlockLastCommitHash) {
		// the BTC validator votes for a fork!

		// construct and save evidence
		evidence := &types.Evidence{
			ValBtcPk:            req.ValBtcPk,
			BlockHeight:         req.BlockHeight,
			BlockLastCommitHash: req.BlockLastCommitHash,
			FinalitySig:         req.FinalitySig,
		}
		ms.SetEvidence(ctx, evidence)

		// if this BTC validator has also signed canonical block, extract its secret key and emit an event
		canonicalSig, err := ms.GetSig(ctx, req.BlockHeight, valPK)
		if err == nil {
			btcSK, err := evidence.ExtractBTCSK(indexedBlock, pubRand, canonicalSig)
			if err != nil {
				panic(fmt.Errorf("failed to extract secret key from two EOTS signatures with the same public randomness: %v", err))
			}

			eventSlashing := types.NewEventSlashedBTCValidator(req.ValBtcPk, indexedBlock, evidence, btcSK)
			if err := ctx.EventManager().EmitTypedEvent(eventSlashing); err != nil {
				panic(fmt.Errorf("failed to emit EventSlashedBTCValidator event: %w", err))
			}
		}

		// NOTE: we should NOT return error here, otherwise the state change triggered in this tx
		// (including the evidence) will be rolled back
		return &types.MsgAddFinalitySigResponse{}, nil
	}

	// this signature is good, add vote to DB
	ms.SetSig(ctx, req.BlockHeight, valPK, req.FinalitySig)

	// if this BTC validator has signed the canonical block before,
	// slash it via extracting its secret key, and emit an event
	if ms.HasEvidence(ctx, req.BlockHeight, req.ValBtcPk) {
		// the BTC validator has voted for a fork before!

		// get evidence
		evidence, err := ms.GetEvidence(ctx, req.BlockHeight, req.ValBtcPk)
		if err != nil {
			panic(fmt.Errorf("failed to get evidence despite HasEvidence returns true"))
		}

		// extract its SK
		btcSK, err := evidence.ExtractBTCSK(indexedBlock, pubRand, req.FinalitySig)
		if err != nil {
			panic(fmt.Errorf("failed to extract secret key from two EOTS signatures with the same public randomness: %v", err))
		}

		eventSlashing := types.NewEventSlashedBTCValidator(req.ValBtcPk, indexedBlock, evidence, btcSK)
		if err := ctx.EventManager().EmitTypedEvent(eventSlashing); err != nil {
			panic(fmt.Errorf("failed to emit EventSlashedBTCValidator event: %w", err))
		}
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
