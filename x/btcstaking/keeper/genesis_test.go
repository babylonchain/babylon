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
	r := rand.New(rand.NewSource(10))
	h := helper.NewHelper(t)

	k, ctx := h.App.BTCStakingKeeper, h.Ctx

	numFps := 3

	fps := datagen.CreateNFinalityProviders(r, t, numFps)
	covQuorum := k.GetParams(h.Ctx).CovenantQuorum

	vpFps := make(map[string]*types.VotingPowerFP, 0)
	btcDelegations := make([]*types.BTCDelegation, 0)
	for _, fp := range fps {
		h.AddFinalityProvider(fp)
		stakingValue := r.Int31n(200000) + 10000
		numDelegations := r.Int31n(10)
		delegations := createNDelegationsForFinalityProvider(
			r,
			t,
			fp.BtcPk.MustToBTCPK(),
			int64(stakingValue),
			int(numDelegations),
			covQuorum,
		)
		blkHeight := r.Uint64()
		vp := uint64(stakingValue)
		k.SetVotingPower(ctx, *fp.BtcPk, blkHeight, vp)

		vpFps[fp.BtcPk.MarshalHex()] = &types.VotingPowerFP{
			BlockHeight: blkHeight,
			FpBtcPk:     fp.BtcPk,
			VotingPower: vp,
		}
		for _, del := range delegations {
			h.AddDelegation(del)
			btcDelegations = append(btcDelegations, del)
		}
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

	for _, gsFpVp := range gs.VotingPowers {
		vp := vpFps[gsFpVp.FpBtcPk.MarshalHex()]
		require.Equal(t, gsFpVp, vp)
	}
}
