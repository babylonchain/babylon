package keeper_test

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/testutil/helper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/stretchr/testify/require"
)

func TestExportGenesis(t *testing.T) {
	r, h := rand.New(rand.NewSource(10)), helper.NewHelper(t)
	k, btclcK, ctx := h.App.BTCStakingKeeper, h.App.BTCLightClientKeeper, h.Ctx
	numFps := 3

	fps := datagen.CreateNFinalityProviders(r, t, numFps)
	params := k.GetParams(ctx)

	chainsHeight := make([]*types.BlockHeightBbnToBtc, 0)
	// creates the first as it starts already with an chain height from the helper.
	chainsHeight = append(chainsHeight, &types.BlockHeightBbnToBtc{
		BlockHeightBbn: 1,
		BlockHeightBtc: 0,
	})
	vpFps := make(map[string]*types.VotingPowerFP, 0)
	btcDelegations := make([]*types.BTCDelegation, 0)
	btcDelegatorIndex := make(map[string]*types.BTCDelegator, 0)

	blkHeight := uint64(r.Int63n(1000))
	btcHead := btclcK.GetTipInfo(ctx)
	totalDelegations := 0

	for _, fp := range fps {
		// set finality
		h.AddFinalityProvider(fp)

		stakingValue := r.Int31n(200000) + 10000
		numDelegations := r.Int31n(10)
		totalDelegations += int(numDelegations)
		delegations := createNDelegationsForFinalityProvider(
			r,
			t,
			fp.BtcPk.MustToBTCPK(),
			int64(stakingValue),
			int(numDelegations),
			params.CovenantQuorum,
		)
		blkHeight++
		vp := uint64(stakingValue)

		// sets voting power
		k.SetVotingPower(ctx, *fp.BtcPk, blkHeight, vp)
		vpFps[fp.BtcPk.MarshalHex()] = &types.VotingPowerFP{
			BlockHeight: blkHeight,
			FpBtcPk:     fp.BtcPk,
			VotingPower: vp,
		}

		for _, del := range delegations {
			// sets delegations
			h.AddDelegation(del)
			btcDelegations = append(btcDelegations, del)

			// delegators

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

		}

		// sets chain heights
		k.IndexBTCHeight(ctx.WithBlockHeight(int64(blkHeight)))
		chainsHeight = append(chainsHeight, &types.BlockHeightBbnToBtc{
			BlockHeightBbn: blkHeight,
			BlockHeightBtc: btcHead.Height,
		})

	}

	gs, err := k.ExportGenesis(ctx)
	h.NoError(err)
	require.Equal(t, k.GetParams(ctx), gs.Params)

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
	// vp dst cache
}
