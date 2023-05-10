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
func (k Keeper) getFinalizedInfo(ctx sdk.Context) (*epochingtypes.Epoch, *checkpointingtypes.RawCheckpoint, *btcctypes.SubmissionKey, *types.ProofEpochSealed, []*btcctypes.TransactionInfo, []*btclctypes.BTCHeaderInfo, error) {
	// get the last finalised epoch metadata
	finalizedEpoch, err := k.GetFinalizedEpoch(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	finalizedEpochInfo, err := k.epochingKeeper.GetHistoricalEpoch(ctx, finalizedEpoch)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// assign raw checkpoint
	rawCheckpoint, err := k.checkpointingKeeper.GetRawCheckpoint(ctx, finalizedEpoch)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// assign BTC submission key
	_, btcSubmissionKey, err := k.btccKeeper.GetBestSubmission(ctx, finalizedEpoch)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// proof that the epoch is sealed
	proofEpochSealed, err := k.ProveEpochSealed(ctx, finalizedEpochInfo.EpochNumber)
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
	oldBTCTip, err := k.GetFinalizingBTCTip(ctx) // NOTE: BTC tip in KVStore has not been updated yet
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	curBTCTip := k.btclcKeeper.GetTipInfo(ctx)
	commonAncestor := k.btclcKeeper.GetHighestCommonAncestor(ctx, oldBTCTip, curBTCTip)
	btcHeaders := k.btclcKeeper.GetInOrderAncestorsUntil(ctx, curBTCTip, commonAncestor)

	return finalizedEpochInfo, rawCheckpoint.Ckpt, btcSubmissionKey, proofEpochSealed, proofEpochSubmitted, btcHeaders, nil
}

// BroadcastBTCTimestamps sends an IBC packet of BTC timestamp to all open IBC channels to ZoneConcierge
func (k Keeper) BroadcastBTCTimestamps(ctx sdk.Context) {
	// get all channels that are open and are connected to ZoneConcierge's port
	openZCChannels := k.GetAllOpenZCChannels(ctx)
	if len(openZCChannels) == 0 {
		k.Logger(ctx).Info("no open IBC channel with ZoneConcierge, skip broadcasting BTC timestamps")
		return
	}

	// get all metadata shared across BTC timestamps in the same epoch
	finalizedEpochInfo, rawCheckpoint, btcSubmissionKey, proofEpochSealed, proofEpochSubmitted, btcHeaders, err := k.getFinalizedInfo(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to generate metadata shared across BTC timestamps in the same epoch", "error", err)
		return
	}

	// for each channel, construct and send BTC timestamp
	for _, channel := range openZCChannels {
		// get the ID of the chain under this channel
		chainID, err := k.getChainID(ctx, channel)
		if err != nil {
			k.Logger(ctx).Error("failed to get chain ID", "channelID", channel.ChannelId, "error", err)
			continue
		}

		// get finalised chainInfo
		finalizedChainInfo, err := k.GetEpochChainInfo(ctx, chainID, finalizedEpochInfo.EpochNumber)
		if err != nil {
			k.Logger(ctx).Error("failed to get finalizedChainInfo", "error", err)
			continue
		}

		// get proofTxInBlock
		proofTxInBlock, err := k.ProveTxInBlock(ctx, finalizedChainInfo.LatestHeader.BabylonTxHash)
		if err != nil {
			k.Logger(ctx).Error("failed to generate proofTxInBlock", "error", err)
			continue
		}

		// get proofHeaderInEpoch
		proofHeaderInEpoch, err := k.ProveHeaderInEpoch(ctx, finalizedChainInfo.LatestHeader.BabylonHeader, finalizedEpochInfo)
		if err != nil {
			k.Logger(ctx).Error("failed to generate proofHeaderInEpoch", "error", err)
			continue
		}

		// construct BTC timestamp from everything
		btcTimestamp := &types.BTCTimestamp{
			Header:           finalizedChainInfo.LatestHeader,
			BtcHeaders:       btcHeaders,
			EpochInfo:        finalizedEpochInfo,
			RawCheckpoint:    rawCheckpoint,
			BtcSubmissionKey: btcSubmissionKey,
			Proof: &types.ProofFinalizedChainInfo{
				ProofTxInBlock:      proofTxInBlock,
				ProofHeaderInEpoch:  proofHeaderInEpoch,
				ProofEpochSealed:    proofEpochSealed,
				ProofEpochSubmitted: proofEpochSubmitted,
			},
		}

		// wrap BTC timestamp to IBC packet
		packet := &types.ZoneconciergePacketData{
			Packet: &types.ZoneconciergePacketData_BtcTimestamp{
				BtcTimestamp: btcTimestamp,
			},
		}

		// if the Babylon contract in this channel has not been initialised, send initialise message first
		if k.channelExists(ctx, channel.ChannelId) {
			if err := k.SendInitBTCHeaders(ctx, channel); err != nil {
				k.Logger(ctx).Error("failed to send InitBTCHeaders IBC packet", "channelID", channel.ChannelId, "error", err)
				continue
			}
			k.afterChannelInited(ctx, channel.ChannelId)
		}

		// send IBC packet
		if err := k.SendIBCPacket(ctx, channel, packet); err != nil {
			k.Logger(ctx).Error("failed to send BTC timestamp IBC packet", "chainID", chainID, "channelID", channel.ChannelId, "error", err)
			continue
		}
	}
}
