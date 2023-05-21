package keeper

import (
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

// finalizedInfo is a private struct that stores metadata and proofs
// identical to all BTC timestamps in the same epoch
type finalizedInfo struct {
	EpochInfo           *epochingtypes.Epoch
	RawCheckpoint       *checkpointingtypes.RawCheckpoint
	BTCSubmissionKey    *btcctypes.SubmissionKey
	ProofEpochSealed    *types.ProofEpochSealed
	ProofEpochSubmitted []*btcctypes.TransactionInfo
	BTCHeaders          []*btclctypes.BTCHeaderInfo
}

// getChainID gets the ID of the counterparty chain under the given channel
func (k Keeper) getChainID(ctx sdk.Context, channel channeltypes.IdentifiedChannel) (string, error) {
	// get clientState under this channel
	_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, channel.PortId, channel.ChannelId)
	if err != nil {
		return "", err
	}
	// cast clientState to tendermint clientState
	tmClient, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return "", fmt.Errorf("client must be a Tendermint client, expected: %T, got: %T", &ibctmtypes.ClientState{}, tmClient)
	}
	return tmClient.ChainId, nil
}

// getBTCHeadersDuringLastFinalizedEpoch gets BTC headers between
// - the block AFTER the common ancestor of BTC tip at epoch `lastFinalizedEpoch-1` and BTC tip at epoch `lastFinalizedEpoch`
// - BTC tip at epoch `lastFinalizedEpoch`
// where `lastFinalizedEpoch` is the last finalised epoch
func (k Keeper) getBTCHeadersDuringLastFinalizedEpoch(ctx sdk.Context) []*btclctypes.BTCHeaderInfo {
	oldBTCTip := k.GetFinalizingBTCTip(ctx) // NOTE: BTC tip in KVStore has not been updated yet
	if oldBTCTip == nil {
		// this happens upon the first finalised epoch. Use base header instead
		oldBTCTip = k.btclcKeeper.GetBaseBTCHeader(ctx)
	}
	curBTCTip := k.btclcKeeper.GetTipInfo(ctx)
	commonAncestor := k.btclcKeeper.GetHighestCommonAncestor(ctx, oldBTCTip, curBTCTip)
	btcHeaders := k.btclcKeeper.GetInOrderAncestorsUntil(ctx, curBTCTip, commonAncestor)

	return btcHeaders
}

// getFinalizedInfo returns metadata and proofs that are identical to all BTC timestamps in the same epoch
func (k Keeper) getFinalizedInfo(ctx sdk.Context, epochNum uint64) (*finalizedInfo, error) {
	finalizedEpochInfo, err := k.epochingKeeper.GetHistoricalEpoch(ctx, epochNum)
	if err != nil {
		return nil, err
	}

	// assign raw checkpoint
	rawCheckpoint, err := k.checkpointingKeeper.GetRawCheckpoint(ctx, epochNum)
	if err != nil {
		return nil, err
	}

	// assign BTC submission key
	ed := k.btccKeeper.GetEpochData(ctx, epochNum)
	bestSubmissionBtcInfo := k.btccKeeper.GetEpochBestSubmissionBtcInfo(ctx, ed)
	if bestSubmissionBtcInfo == nil {
		return nil, fmt.Errorf("empty bestSubmissionBtcInfo")
	}
	btcSubmissionKey := &bestSubmissionBtcInfo.SubmissionKey

	// proof that the epoch is sealed
	proofEpochSealed, err := k.ProveEpochSealed(ctx, epochNum)
	if err != nil {
		return nil, err
	}

	// proof that the epoch's checkpoint is submitted to BTC
	// i.e., the two `TransactionInfo`s for the checkpoint
	proofEpochSubmitted, err := k.ProveEpochSubmitted(ctx, btcSubmissionKey)
	if err != nil {
		return nil, err
	}

	// get new BTC headers since the 2nd last finalised epoch and the last finalised epoch
	btcHeaders := k.getBTCHeadersDuringLastFinalizedEpoch(ctx)

	// construct finalizedInfo
	finalizedInfo := &finalizedInfo{
		EpochInfo:           finalizedEpochInfo,
		RawCheckpoint:       rawCheckpoint.Ckpt,
		BTCSubmissionKey:    btcSubmissionKey,
		ProofEpochSealed:    proofEpochSealed,
		ProofEpochSubmitted: proofEpochSubmitted,
		BTCHeaders:          btcHeaders,
	}

	return finalizedInfo, nil
}

// createBTCTimestamp creates a BTC timestamp from finalizedInfo for a given IBC channel
// where the counterparty is a Cosmos zone
func (k Keeper) createBTCTimestamp(ctx sdk.Context, chainID string, channelID string, finalizedInfo *finalizedInfo) (*types.BTCTimestamp, error) {
	// if the Babylon contract in this channel has not been initialised, get headers from
	// the tip to (w+1+len(finalizedInfo.BTCHeaders))-deep header
	var btcHeaders []*btclctypes.BTCHeaderInfo
	if k.isChannelUninitialized(ctx, channelID) {
		w := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout
		depth := w + 1 + uint64(len(finalizedInfo.BTCHeaders))

		btcHeaders = k.btclcKeeper.GetMainChainUpTo(ctx, depth)
		if btcHeaders == nil {
			return nil, fmt.Errorf("failed to get Bitcoin main chain up to depth %d", depth)
		}
		bbn.Reverse(btcHeaders)
	} else {
		btcHeaders = finalizedInfo.BTCHeaders
	}

	// get finalised chainInfo
	// NOTE: it's possible that this chain does not have chain info at the moment
	// In this case, skip sending BTC timestamp for this chain at this epoch
	epochNum := finalizedInfo.EpochInfo.EpochNumber
	finalizedChainInfo, err := k.GetEpochChainInfo(ctx, chainID, epochNum)
	if err != nil {
		return nil, fmt.Errorf("no finalizedChainInfo for chain %s at epoch %d", chainID, epochNum)
	}

	// construct BTC timestamp from everything
	// NOTE: it's possible that there is no header checkpointed in this epoch
	btcTimestamp := &types.BTCTimestamp{
		Header:           nil,
		BtcHeaders:       btcHeaders,
		EpochInfo:        finalizedInfo.EpochInfo,
		RawCheckpoint:    finalizedInfo.RawCheckpoint,
		BtcSubmissionKey: finalizedInfo.BTCSubmissionKey,
		Proof: &types.ProofFinalizedChainInfo{
			ProofTxInBlock:      nil,
			ProofHeaderInEpoch:  nil,
			ProofEpochSealed:    finalizedInfo.ProofEpochSealed,
			ProofEpochSubmitted: finalizedInfo.ProofEpochSubmitted,
		},
	}

	// if there is a CZ header checkpointed in this finalised epoch,
	// add this CZ header and corresponding proofs to the BTC timestamp
	if finalizedChainInfo.LatestHeader.BabylonEpoch == epochNum {
		// get proofTxInBlock
		proofTxInBlock, err := k.ProveTxInBlock(ctx, finalizedChainInfo.LatestHeader.BabylonTxHash)
		if err != nil {
			return nil, fmt.Errorf("failed to generate proofTxInBlock for chain %s: %w", chainID, err)
		}

		// get proofHeaderInEpoch
		proofHeaderInEpoch, err := k.ProveHeaderInEpoch(ctx, finalizedChainInfo.LatestHeader.BabylonHeader, finalizedInfo.EpochInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate proofHeaderInEpoch for chain %s: %w", chainID, err)
		}

		btcTimestamp.Header = finalizedChainInfo.LatestHeader
		btcTimestamp.Proof.ProofTxInBlock = proofTxInBlock
		btcTimestamp.Proof.ProofHeaderInEpoch = proofHeaderInEpoch
	}

	return btcTimestamp, nil
}

// BroadcastBTCTimestamps sends an IBC packet of BTC timestamp to all open IBC channels to ZoneConcierge
func (k Keeper) BroadcastBTCTimestamps(ctx sdk.Context, epochNum uint64) {
	// Babylon does not broadcast BTC timestamps until finalising epoch 1
	if epochNum < 1 {
		k.Logger(ctx).Info("Babylon does not finalize epoch 1 yet, skip broadcasting BTC timestamps")
		return
	}

	// get all channels that are open and are connected to ZoneConcierge's port
	openZCChannels := k.GetAllOpenZCChannels(ctx)
	if len(openZCChannels) == 0 {
		k.Logger(ctx).Info("no open IBC channel with ZoneConcierge, skip broadcasting BTC timestamps")
		return
	}

	k.Logger(ctx).Info("there exists open IBC channels with ZoneConcierge, generating BTC timestamps", "number of channels", len(openZCChannels))

	// get all metadata shared across BTC timestamps in the same epoch
	finalizedInfo, err := k.getFinalizedInfo(ctx, epochNum)
	if err != nil {
		k.Logger(ctx).Error("failed to generate metadata shared across BTC timestamps in the same epoch, skip broadcasting BTC timestamps", "error", err)
		return
	}

	// for each channel, construct and send BTC timestamp
	for _, channel := range openZCChannels {
		// get the ID of the chain under this channel
		chainID, err := k.getChainID(ctx, channel)
		if err != nil {
			k.Logger(ctx).Error("failed to get chain ID, skip sending BTC timestamp for this chain", "channelID", channel.ChannelId, "error", err)
			continue
		}

		// generate timestamp for this channel
		btcTimestamp, err := k.createBTCTimestamp(ctx, chainID, channel.ChannelId, finalizedInfo)
		if err != nil {
			k.Logger(ctx).Error("failed to generate BTC timestamp, skip sending BTC timestamp for this chain", "chainID", chainID, "error", err)
			continue
		}

		// wrap BTC timestamp to IBC packet
		packet := types.NewBTCTimestampPacketData(btcTimestamp)
		// send IBC packet
		if err := k.SendIBCPacket(ctx, channel, packet); err != nil {
			k.Logger(ctx).Error("failed to send BTC timestamp IBC packet, skip sending BTC timestamp for this chain", "chainID", chainID, "channelID", channel.ChannelId, "error", err)
			continue
		}

		// this channel has been initialised after sending the first IBC packet
		if k.isChannelUninitialized(ctx, channel.ChannelId) {
			k.afterChannelInitialized(ctx, channel.ChannelId)
		}
	}
}

// TODO: test case with at BTC headers and checkpoints
