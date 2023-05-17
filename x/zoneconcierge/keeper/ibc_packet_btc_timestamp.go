package keeper

import (
	"fmt"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

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

// getFinalizedInfo returns metadata and proofs that are identical to all BTC timestamps in the same epoch
func (k Keeper) getFinalizedInfo(ctx sdk.Context, epochNum uint64) (*epochingtypes.Epoch, *checkpointingtypes.RawCheckpoint, *btcctypes.SubmissionKey, *types.ProofEpochSealed, []*btcctypes.TransactionInfo, []*btclctypes.BTCHeaderInfo, error) {
	finalizedEpochInfo, err := k.epochingKeeper.GetHistoricalEpoch(ctx, epochNum)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// assign raw checkpoint
	rawCheckpoint, err := k.checkpointingKeeper.GetRawCheckpoint(ctx, epochNum)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// assign BTC submission key
	_, btcSubmissionKey, err := k.btccKeeper.GetBestSubmission(ctx, epochNum)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// proof that the epoch is sealed
	proofEpochSealed, err := k.ProveEpochSealed(ctx, epochNum)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// proof that the epoch's checkpoint is submitted to BTC
	// i.e., the two `TransactionInfo`s for the checkpoint
	proofEpochSubmitted, err := k.ProveEpochSubmitted(ctx, btcSubmissionKey)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// Get BTC headers between
	// - the block AFTER the common ancestor of BTC tip at epoch `lastFinalizedEpoch-1` and BTC tip at epoch `lastFinalizedEpoch`
	// - BTC tip at epoch `lastFinalizedEpoch`
	oldBTCTip := k.GetFinalizingBTCTip(ctx) // NOTE: BTC tip in KVStore has not been updated yet
	if oldBTCTip == nil {
		// this happens upon the first finalised epoch. Use base header instead
		oldBTCTip = k.btclcKeeper.GetBaseBTCHeader(ctx)
	}
	curBTCTip := k.btclcKeeper.GetTipInfo(ctx)
	commonAncestor := k.btclcKeeper.GetHighestCommonAncestor(ctx, oldBTCTip, curBTCTip)
	btcHeaders := k.btclcKeeper.GetInOrderAncestorsUntil(ctx, curBTCTip, commonAncestor)

	return finalizedEpochInfo, rawCheckpoint.Ckpt, btcSubmissionKey, proofEpochSealed, proofEpochSubmitted, btcHeaders, nil
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
	finalizedEpochInfo, rawCheckpoint, btcSubmissionKey, proofEpochSealed, proofEpochSubmitted, epochBtcHeaders, err := k.getFinalizedInfo(ctx, epochNum)
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

		k.Logger(ctx).Info("sending BTC timestamp to channel", "channelID", channel.ChannelId, "chainID", chainID)

		// if the Babylon contract in this channel has not been initialised, prepend w headers as w-deep proof
		var btcHeaders []*btclctypes.BTCHeaderInfo
		if k.isChannelUninitialized(ctx, channel.ChannelId) {
			w := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout
			prependingHeaders, err := k.btclcKeeper.GetInOrderAncestorsUntilHeight(ctx, w, epochBtcHeaders[0].Height-1)
			if err != nil {
				k.Logger(ctx).Error("failed to get w+1 headers, skip sending BTC timestamp for this chain", "chainID", chainID, "error", err)
				continue
			}
			btcHeaders = append(prependingHeaders, epochBtcHeaders...)
		} else {
			btcHeaders = epochBtcHeaders
		}

		// get finalised chainInfo
		// NOTE: it's possible that this chain does not have chain info at the moment
		// In this case, skip sending BTC timestamp for this chain at this epoch
		finalizedChainInfo, err := k.GetEpochChainInfo(ctx, chainID, epochNum)
		if err != nil {
			k.Logger(ctx).Info("no finalizedChainInfo for this chain at this epoch, skip sending BTC timestamp for this chain", "chainID", chainID, "epoch", epochNum, "error", err)
			continue
		}

		// construct BTC timestamp from everything
		// NOTE: it's possible that there is no header checkpointed in this epoch
		btcTimestamp := &types.BTCTimestamp{
			Header:           nil,
			BtcHeaders:       btcHeaders,
			EpochInfo:        finalizedEpochInfo,
			RawCheckpoint:    rawCheckpoint,
			BtcSubmissionKey: btcSubmissionKey,
			Proof: &types.ProofFinalizedChainInfo{
				ProofTxInBlock:      nil,
				ProofHeaderInEpoch:  nil,
				ProofEpochSealed:    proofEpochSealed,
				ProofEpochSubmitted: proofEpochSubmitted,
			},
		}

		// if there is a CZ header checkpointed in this finalised epoch,
		// add this CZ header and corresponding proofs to the BTC timestamp
		if finalizedChainInfo.LatestHeader.BabylonEpoch == epochNum {
			// get proofTxInBlock
			proofTxInBlock, err := k.ProveTxInBlock(ctx, finalizedChainInfo.LatestHeader.BabylonTxHash)
			if err != nil {
				k.Logger(ctx).Error("failed to generate proofTxInBlock, skip sending BTC timestamp for this chain", "chainID", chainID, "error", err)
				continue
			}

			// get proofHeaderInEpoch
			proofHeaderInEpoch, err := k.ProveHeaderInEpoch(ctx, finalizedChainInfo.LatestHeader.BabylonHeader, finalizedEpochInfo)
			if err != nil {
				k.Logger(ctx).Error("failed to generate proofHeaderInEpoch, skip sending BTC timestamp for this chain", "chainID", chainID, "error", err)
				continue
			}

			btcTimestamp.Header = finalizedChainInfo.LatestHeader
			btcTimestamp.Proof.ProofTxInBlock = proofTxInBlock
			btcTimestamp.Proof.ProofHeaderInEpoch = proofHeaderInEpoch
		}

		// wrap BTC timestamp to IBC packet
		packet := &types.ZoneconciergePacketData{
			Packet: &types.ZoneconciergePacketData_BtcTimestamp{
				BtcTimestamp: btcTimestamp,
			},
		}

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
