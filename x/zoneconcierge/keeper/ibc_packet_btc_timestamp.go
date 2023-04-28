package keeper

import (
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
		// The only error in ProveEpochSubmitted is the nil bestSubmission.
		// Since the epoch w.r.t. the bestSubmissionKey is finalised, this
		// can only be a programming error, so we should panic here.
		panic(err)
	}

	// TODO: get BTC headers between
	// - the common ancestor of BTC tip at epoch `lastFinalizedEpoch-1` and BTC tip at epoch `lastFinalizedEpoch`
	// - BTC tip at epoch `lastFinalizedEpoch`
	btcHeaders := []*btclctypes.BTCHeaderInfo{}

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
		// TODO: find a way to get chain ID from channel ID
		chainID := "fixme"

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

		// send IBC packet
		if err := k.SendIBCPacket(ctx, channel, packet); err != nil {
			k.Logger(ctx).Error("failed to send BTC timestamp IBC packet", "chainID", chainID, "channelID", channel.ChannelId, "error", err)
			continue
		}
	}
}
