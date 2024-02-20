package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzFinalityProviderEvents(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set all parameters
		h.GenAndApplyParams(r)

		// generate and insert new finality provider
		_, _, fp := h.CreateFinalityProvider(r)

		// mock BTC tip info
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()

		// slash the finality provider
		err := h.BTCStakingKeeper.SlashFinalityProvider(h.Ctx, fp.BtcPk.MustMarshal())
		h.NoError(err)

		// at this point, there should be only 1 event that the finality provider is slashed
		btcTipHeight := btclcKeeper.GetTipInfo(h.Ctx).Height
		h.BTCStakingKeeper.IteratePowerDistUpdateEvents(h.Ctx, btcTipHeight, func(ev *types.EventPowerDistUpdate) bool {
			slashedFPEvent := ev.GetSlashedFp()
			require.NotNil(t, slashedFPEvent)
			require.Equal(t, fp.BtcPk.MustMarshal(), slashedFPEvent.Pk.MustMarshal())
			return true
		})
	})
}

func FuzzBTCDelegationEvents(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set all parameters
		covenantSKs, _ := h.GenAndApplyParams(r)
		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		require.NoError(t, err)

		// generate and insert new finality provider
		_, fpPK, _ := h.CreateFinalityProvider(r)

		// generate and insert new BTC delegation
		stakingValue := int64(2 * 10e8)
		expectedStakingTxHash, _, _, msgCreateBTCDel, actualDel := h.CreateDelegation(
			r,
			fpPK,
			changeAddress.EncodeAddress(),
			stakingValue,
			1000,
		)

		/*
			at this point, there should be
			- 1 event that BTC delegation becomes pending at current BTC tip
			- 1 event that BTC delegation will become expired at end height - w
		*/
		// the BTC delegation is now pending
		btcTipHeight := btclcKeeper.GetTipInfo(h.Ctx).Height
		h.BTCStakingKeeper.IteratePowerDistUpdateEvents(h.Ctx, btcTipHeight, func(ev *types.EventPowerDistUpdate) bool {
			btcDelStateUpdate := ev.GetBtcDelStateUpdate()
			require.NotNil(t, btcDelStateUpdate)
			require.Equal(t, expectedStakingTxHash, btcDelStateUpdate.StakingTxHash)
			require.Equal(t, types.BTCDelegationStatus_PENDING, btcDelStateUpdate.NewState)
			return true
		})
		// the BTC delegation will be unbonded at end height - w
		unbondedHeight := actualDel.EndHeight - btccKeeper.GetParams(h.Ctx).CheckpointFinalizationTimeout
		h.BTCStakingKeeper.IteratePowerDistUpdateEvents(h.Ctx, unbondedHeight, func(ev *types.EventPowerDistUpdate) bool {
			btcDelStateUpdate := ev.GetBtcDelStateUpdate()
			require.NotNil(t, btcDelStateUpdate)
			require.Equal(t, expectedStakingTxHash, btcDelStateUpdate.StakingTxHash)
			require.Equal(t, types.BTCDelegationStatus_UNBONDED, btcDelStateUpdate.NewState)
			return true
		})

		// clear the events at tip, as per the behaviour of each `BeginBlock`
		h.BTCStakingKeeper.ClearPowerDistUpdateEvents(h.Ctx, btcTipHeight)
		// ensure event queue is cleared at BTC tip height
		numEvents := 0
		h.BTCStakingKeeper.IteratePowerDistUpdateEvents(h.Ctx, btcTipHeight, func(ev *types.EventPowerDistUpdate) bool {
			numEvents++
			return true
		})
		require.Zero(t, numEvents)

		// generate a quorum number of covenant signatures
		msgs := h.GenerateCovenantSignaturesMessages(r, covenantSKs, msgCreateBTCDel, actualDel)
		for i := 0; i < int(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum); i++ {
			_, err = h.MsgServer.AddCovenantSigs(h.Ctx, msgs[i])
			h.NoError(err)
		}

		/*
			at this point, there should be an event that the BTC delegation becomes
			active at the current height
		*/
		btcTipHeight = btclcKeeper.GetTipInfo(h.Ctx).Height
		h.BTCStakingKeeper.IteratePowerDistUpdateEvents(h.Ctx, btcTipHeight, func(ev *types.EventPowerDistUpdate) bool {
			btcDelStateUpdate := ev.GetBtcDelStateUpdate()
			require.NotNil(t, btcDelStateUpdate)
			require.Equal(t, expectedStakingTxHash, btcDelStateUpdate.StakingTxHash)
			require.Equal(t, types.BTCDelegationStatus_ACTIVE, btcDelStateUpdate.NewState)
			return true
		})
	})
}
