package keeper_test

import (
	"bytes"
	"math"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/testutil/helper"
	btclightclientt "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/stretchr/testify/require"
)

func TestExportGenesis(t *testing.T) {
	r, h := rand.New(rand.NewSource(11)), helper.NewHelper(t)
	k, btclcK, btcCheckK, ctx := h.App.BTCStakingKeeper, h.App.BTCLightClientKeeper, h.App.BtcCheckpointKeeper, h.Ctx
	numFps := 3

	fps := datagen.CreateNFinalityProviders(r, t, numFps)
	params := k.GetParams(ctx)
	wValue := btcCheckK.GetParams(ctx).CheckpointFinalizationTimeout

	chainsHeight := make([]*types.BlockHeightBbnToBtc, 0)
	// creates the first as it starts already with an chain height from the helper.
	chainsHeight = append(chainsHeight, &types.BlockHeightBbnToBtc{
		BlockHeightBbn: 1,
		BlockHeightBtc: 0,
	})
	vpFps := make(map[string]*types.VotingPowerFP, 0)
	btcDelegations := make([]*types.BTCDelegation, 0)
	eventsIdx := make(map[uint64]*types.EventIndex, 0)
	btcDelegatorIndex := make(map[string]*types.BTCDelegator, 0)

	blkHeight := uint64(r.Int63n(1000)) + math.MaxUint16
	totalDelegations := 0

	for _, fp := range fps {
		btcHead := btclcK.GetTipInfo(ctx)
		btcHead.Height = blkHeight + 100
		btclcK.InsertHeaderInfos(ctx, []*btclightclientt.BTCHeaderInfo{
			btcHead,
		})

		// set finality
		h.AddFinalityProvider(fp)

		stakingValue := r.Int31n(200000) + 10000
		numDelegations := r.Int31n(10)
		delegations := createNDelegationsForFinalityProvider(
			r,
			t,
			fp.BtcPk.MustToBTCPK(),
			int64(stakingValue),
			int(numDelegations),
			params.CovenantQuorum,
		)
		vp := uint64(stakingValue)

		// sets voting power
		k.SetVotingPower(ctx, *fp.BtcPk, blkHeight, vp)
		vpFps[fp.BtcPk.MarshalHex()] = &types.VotingPowerFP{
			BlockHeight: blkHeight,
			FpBtcPk:     fp.BtcPk,
			VotingPower: vp,
		}

		for _, del := range delegations {
			totalDelegations++

			// sets delegations
			h.AddDelegation(del)
			btcDelegations = append(btcDelegations, del)

			// BTC delegators idx
			stakingTxHash, err := del.GetStakingTxHash()
			h.NoError(err)

			idxDelegatorStk := types.NewBTCDelegatorDelegationIndex()
			err = idxDelegatorStk.Add(stakingTxHash)
			h.NoError(err)

			btcDelegatorIndex[del.BtcPk.MarshalHex()] = &types.BTCDelegator{
				Idx: &types.BTCDelegatorDelegationIndex{
					StakingTxHashList: idxDelegatorStk.StakingTxHashList,
				},
				FpBtcPk:  fp.BtcPk,
				DelBtcPk: del.BtcPk,
			}

			// record event that the BTC delegation will become unbonded at endHeight-w
			unbondedEvent := types.NewEventPowerDistUpdateWithBTCDel(&types.EventBTCDelegationStateUpdate{
				StakingTxHash: stakingTxHash.String(),
				NewState:      types.BTCDelegationStatus_UNBONDED,
			})

			// events
			idxEvent := uint64(totalDelegations - 1)
			eventsIdx[idxEvent] = &types.EventIndex{
				Idx:            idxEvent,
				BlockHeightBtc: del.EndHeight - wValue,
				Event:          unbondedEvent,
			}
		}

		// sets chain heights
		header := ctx.HeaderInfo()
		header.Height = int64(blkHeight)
		ctx = ctx.WithHeaderInfo(header)
		h.Ctx = ctx

		k.IndexBTCHeight(ctx)
		chainsHeight = append(chainsHeight, &types.BlockHeightBbnToBtc{
			BlockHeightBbn: blkHeight,
			BlockHeightBtc: btcHead.Height,
		})

		blkHeight++ // each fp increase blk height to modify data in state.
	}

	gs, err := k.ExportGenesis(ctx)
	h.NoError(err)
	require.Equal(t, k.GetParams(ctx), *gs.Params[0])

	// finality providers
	correctFps := 0
	for _, fp := range fps {
		for _, gsfp := range gs.FinalityProviders {
			if !bytes.Equal(fp.BabylonPk.Address(), gsfp.BabylonPk.Address()) {
				continue
			}
			require.EqualValues(t, fp, gsfp)
			correctFps++
		}
	}
	require.Equal(t, correctFps, numFps)

	// btc delegations
	correctDels := 0
	for _, del := range btcDelegations {
		for _, gsdel := range gs.BtcDelegations {
			if !bytes.Equal(del.BabylonPk.Address(), gsdel.BabylonPk.Address()) {
				continue
			}
			correctDels++
			require.Equal(t, del, gsdel)
		}
	}
	require.Equal(t, correctDels, len(btcDelegations))

	// voting powers
	for _, gsFpVp := range gs.VotingPowers {
		vp := vpFps[gsFpVp.FpBtcPk.MarshalHex()]
		require.Equal(t, gsFpVp, vp)
	}

	// chains height
	require.Equal(t, chainsHeight, gs.BlockHeightChains)

	// btc delegators
	require.Equal(t, totalDelegations, len(gs.BtcDelegators))
	for _, btcDel := range gs.BtcDelegators {
		idxBtcDel := btcDelegatorIndex[btcDel.DelBtcPk.MarshalHex()]
		require.Equal(t, btcDel, idxBtcDel)
	}

	// events
	require.Equal(t, totalDelegations, len(gs.Events))
	for _, evt := range gs.Events {
		evtIdx := eventsIdx[evt.Idx]
		require.Equal(t, evt, evtIdx)
	}

	// TODO: vp dst cache
}
