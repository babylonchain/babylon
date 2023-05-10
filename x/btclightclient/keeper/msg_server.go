package keeper

import (
	"context"
	"math/big"

	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcd/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	// This should be a reference to Keeper
	k Keeper
}

func MsgInsertHeaderWrapped(ctx context.Context, k Keeper, msg *types.MsgInsertHeader,
	powLimit big.Int, reduceMinDifficulty bool, retargetAdjustmentFactor int64, powCheck bool,
) (*types.MsgInsertHeaderResponse, error) {
	// Perform the checks that checkBlockHeaderContext of btcd does
	// https://github.com/btcsuite/btcd/blob/master/blockchain/validate.go#L644
	// We skip the time, checkpoint, and version checks
	// TODO: Implement an AnteHandler that performs these checks
	// 		 so as to not pollute the mempool with transactions
	// 		 that will get rejected.
	if msg == nil {
		return nil, types.ErrEmptyMessage.Wrapf("message is nil")
	}

	if msg.Header == nil {
		return nil, types.ErrEmptyMessage.Wrapf("message header is nit")
	}

	// Get the SDK wrapped context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	parentHash := msg.Header.ParentHash()
	// Retrieve parent
	parent, err := k.headersState(sdkCtx).GetHeaderByHash(parentHash)
	// parent does not exist
	if err != nil {
		return nil, err
	}

	/*
		  Verify the work of the new header.
		  Bitcoin core does this verification at:
		  https://github.com/bitcoin/bitcoin/blob/a688ff9046a9df58a373086445ab5796cccf9dd3/src/validation.cpp#L3468
		  This function is invoked to identify the value that the `Bits` field should have:
		  https://github.com/bitcoin/bitcoin/blob/a688ff9046a9df58a373086445ab5796cccf9dd3/src/pow.cpp#L13

		  **Goal**
		  The goal of this check is to avoid the flooding of the btclightclient with headers that are easy to generate.
		  We want to avoid adding very complex checks here since they can be a source of bugs. Therefore,
		  we are ok getting headers that could not be part of the canonical chain, as long as sufficient
		  work has been put to generate them.

		  The algorithm works as follows:
		  Every `params.DifficultyAdjustmentInterval()` the required work of the header is subject to change
		  in order to help maintain a 10-minutes average time between blocks.

		  The configuration contains a `params.ReduceMinDifficulty` parameter that allows headers
		  to have the minimum amount of work allowed by the network regardless of the previous header's work.
		  For the mainnet this is set to false, while for the testnet/simnet this is set to true.

		  Note: despite the naming, `params.powLimit` refers to the *minimum* work that is allowed.
				 However, due to the format of the Bits field, when converting those to big ints,
				 the following check reveals whether the work is more than the minimum:
					   workBigInt < powLimitBigInt

		  1. If the new block is NOT the first header of the adjustment interval:
			   a. If `params.ReduceMinDifficulty` has been set
					i. If the time of the new block is after 20 minutes from the last block (`params.nPowTargetSpacing*2`)
						  The expected work of the header is `params.powLimit`.
					ii. Otherwise,
						  The expected work of the header is equal to one of its ancestors.
						  The ancestor is selected by traversing all ancestors in the given `params.DifficultyAdjustmentInterval()`
						  and selecting the first one that has either work that is more than the `params.powLimit` or the first one
						  in the interval.
			   b. Otherwise,
					i. The expected work of the header is equal to the one of its parent.

			   From the above and given our goal, a valid check for this case would be that the header
			   has work that is at least more than the `params.powLimit`.
			   We will not check whether the work of the new header is exactly the same as the one that is
			   expected.

		  2. Otherwise,
			   a. Get the first block of the `params.DifficultyAdjustmentInterval()`
			   b. Calculate the timespan between the last block and the first block of the interval.
			   c. Ensure that the expected work won't wildly fluctuate from the work of the parent:
					i. If timespan < `params.nPowTargetTimespan / params.retargetAdjustmentFactor`
						  timespan = `params.nPowTargetTimespan / params.retargetAdjustmentFactor`
					ii. if timespan > `params.nPowTargetTimespan * retarget.AdjustmentFactor`
						  timespan = `params.nPowTargetTimespan * params.retargetAdjustmentFactor`

					For both the mainnet and the testnet/simnet `params.retargetAdjustmentFactor = 4`.
			   d. Get PoW of the last block and calculate:
					newPow = parentPow * timespan / `params.nPowTargetTimespan`

					From the above calculation and based on (c), we can get the property:
					parentPow / `params.retargetAdjustmentFactor` <= newPow <= parentPow * `params.retargetAdjustmentFactor`
			   e. If newPow > `params.PowLimit`
					newPow = `params.PowLimit`

			   From the above and given our goal, a valid check for this case would be that:
					i. The header has work that is at least more than the `params.powLimit`
					ii. The header has work that is between the multiple and dividend of the parent work and `params.retargetAdjustmentFactor`

		  Given the stated goal, we would like to reduce complexity as much as possible. Therefore,
		  here we have decided to not differentiate cases (1) and (2). More specifically, the checks that we do are:
		  1. Always verify that the work is at least more than `params.powLimit`.
		  2. In the case that `params.ReduceMinDifficulty` has been set to `false`, check
			  that the header has work that is between the multiple and dividend of the parent work and `params.retargetAdjustmentFactor`
			  This only happens on the mainnet.

		  The above checks lead to some more clutter on the testnet, since some headers won't do check (2) while they should,
		  but the testnet already allows for minimum work headers so that can be tolerated. For the mainnet,
		  we will do all required checks, but we will still not test the exact value of the `Bits` field.
		  Instead we will verify that the `Bits` field value is going to be in a valid range and if someone
		  does BTC mainnet work to add clutter on the BBN chain, then we can tolerate that.
	*/

	if powCheck {
		msgBlock := &wire.MsgBlock{Header: *(msg.Header.ToBlockHeader())}
		block := btcutil.NewBlock(msgBlock)
		err = blockchain.CheckProofOfWork(block, &powLimit)
		if err != nil {
			return nil, types.ErrInvalidProofOfWOrk
		}
	}

	if !reduceMinDifficulty {
		// The new block will either be the first block of a recalculation event
		// which happens every 2,016 blocks or a normal block.
		// In the second case, it's difficulty should be exactly the same as it's parent
		// while in the second case it should have a maximum difference of a factor of 4 from it
		// See: https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch10.asciidoc#retargeting-to-adjust-difficulty
		// We consolidate those into a single check.
		oldDifficulty := blockchain.CompactToBig(parent.Header.Bits())
		currentDifficulty := blockchain.CompactToBig(msg.Header.Bits())
		maxCurrentDifficulty := new(big.Int).Mul(oldDifficulty, big.NewInt(retargetAdjustmentFactor))
		minCurrentDifficulty := new(big.Int).Div(oldDifficulty, big.NewInt(retargetAdjustmentFactor))
		if currentDifficulty.Cmp(maxCurrentDifficulty) > 0 || currentDifficulty.Cmp(minCurrentDifficulty) < 0 {
			return nil, types.ErrInvalidDifficulty.Wrap("difficulty not relevant to parent difficulty")
		}
	}

	// All good, insert the header
	err = k.InsertHeader(sdkCtx, msg.Header)
	if err != nil {
		return nil, err
	}
	return &types.MsgInsertHeaderResponse{}, nil
}

func (m msgServer) InsertHeader(ctx context.Context, msg *types.MsgInsertHeader) (*types.MsgInsertHeaderResponse, error) {
	return MsgInsertHeaderWrapped(
		ctx,
		m.k,
		msg,
		m.k.btcConfig.PowLimit(),
		m.k.btcConfig.ReduceMinDifficulty(),
		m.k.btcConfig.RetargetAdjustmentFactor(),
		true,
	)
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

var _ types.MsgServer = msgServer{}
