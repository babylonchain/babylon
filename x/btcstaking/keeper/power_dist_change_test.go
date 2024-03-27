package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzProcessAllPowerDistUpdateEvents_Determinism(f *testing.F) {
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
		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		require.NoError(t, err)

		// generate and insert a number of new finality providers
		fpPKs := []*btcec.PublicKey{}
		for i := 0; i < 5; i++ {
			_, fpPK, _ := h.CreateFinalityProvider(r)
			fpPKs = append(fpPKs, fpPK)
		}

		// empty dist cache
		dc := types.NewVotingPowerDistCache()

		stakingValue := int64(2 * 10e8)

		// generate many new BTC delegations under each finality provider, and their corresponding events
		events := []*types.EventPowerDistUpdate{}
		for _, fpPK := range fpPKs {
			for i := 0; i < 5; i++ {
				_, _, _, _, del := h.CreateDelegation(r, fpPK, changeAddress.EncodeAddress(), stakingValue, 1000)
				event := types.NewEventPowerDistUpdateWithBTCDel(&types.EventBTCDelegationStateUpdate{
					StakingTxHash: del.MustGetStakingTxHash().String(),
					NewState:      types.BTCDelegationStatus_ACTIVE,
				})
				events = append(events, event)
			}
		}

		newDc := h.BTCStakingKeeper.ProcessAllPowerDistUpdateEvents(h.Ctx, dc, events, 100)
		for i := 0; i < 10; i++ {
			newDc2 := h.BTCStakingKeeper.ProcessAllPowerDistUpdateEvents(h.Ctx, dc, events, 100)
			require.Equal(t, newDc, newDc2)
		}
	})
}

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
		covenantSKs, _ := h.GenAndApplyParams(r)
		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		require.NoError(t, err)

		// generate and insert new finality provider
		_, fpPK, fp := h.CreateFinalityProvider(r)

		/*
			insert new BTC delegation and give it covenant quorum
			ensure that it has voting power
		*/
		stakingValue := int64(2 * 10e8)
		_, _, _, msgCreateBTCDel, actualDel := h.CreateDelegation(
			r,
			fpPK,
			changeAddress.EncodeAddress(),
			stakingValue,
			1000,
		)
		// give it a quorum number of covenant signatures
		msgs := h.GenerateCovenantSignaturesMessages(r, covenantSKs, msgCreateBTCDel, actualDel)
		for i := 0; i < int(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum); i++ {
			_, err = h.MsgServer.AddCovenantSigs(h.Ctx, msgs[i])
			h.NoError(err)
		}

		// execute BeginBlock
		btcTip := btclcKeeper.GetTipInfo(h.Ctx)
		babylonHeight := datagen.RandomInt(r, 10) + 1
		h.SetCtxHeight(babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(btcTip).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		h.NoError(err)
		// ensure the finality provider has voting power at this height
		require.Equal(t, uint64(stakingValue), h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fp.BtcPk, babylonHeight))

		/*
			Slash the finality provider and execute BeginBlock
			Then, ensure the finality provider does not have voting power anymore
		*/
		err = h.BTCStakingKeeper.SlashFinalityProvider(h.Ctx, fp.BtcPk.MustMarshal())
		h.NoError(err)

		// at this point, there should be only 1 event that the finality provider is slashed
		btcTipHeight := btclcKeeper.GetTipInfo(h.Ctx).Height
		h.BTCStakingKeeper.IteratePowerDistUpdateEvents(h.Ctx, btcTipHeight, func(ev *types.EventPowerDistUpdate) bool {
			slashedFPEvent := ev.GetSlashedFp()
			require.NotNil(t, slashedFPEvent)
			require.Equal(t, fp.BtcPk.MustMarshal(), slashedFPEvent.Pk.MustMarshal())
			return true
		})

		// execute BeginBlock
		babylonHeight += 1
		h.SetCtxHeight(babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(btcTip).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		h.NoError(err)
		// ensure the finality provider does not have voting power anymore
		require.Zero(t, h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fp.BtcPk, babylonHeight))
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
		_, fpPK, fp := h.CreateFinalityProvider(r)

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
			at this point, there should be 1 event that BTC delegation
			will become expired at end height - w
		*/
		// there exists no event at the current BTC tip
		btcTip := btclcKeeper.GetTipInfo(h.Ctx)
		events := h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(h.Ctx, btcTip.Height, btcTip.Height)
		require.Len(t, events, 0)
		// the BTC delegation will be unbonded at end height - w
		unbondedHeight := actualDel.EndHeight - btccKeeper.GetParams(h.Ctx).CheckpointFinalizationTimeout
		events = h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(h.Ctx, unbondedHeight, unbondedHeight)
		require.Len(t, events, 1)
		btcDelStateUpdate := events[0].GetBtcDelStateUpdate()
		require.NotNil(t, btcDelStateUpdate)
		require.Equal(t, expectedStakingTxHash, btcDelStateUpdate.StakingTxHash)
		require.Equal(t, types.BTCDelegationStatus_UNBONDED, btcDelStateUpdate.NewState)

		// ensure this finality provider does not have voting power at the current height
		babylonHeight := datagen.RandomInt(r, 10) + 1
		h.SetCtxHeight(babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(btcTip).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		h.NoError(err)
		require.Zero(t, h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fp.BtcPk, babylonHeight))

		/*
			Generate a quorum number of covenant signatures
			Then, there should be an event that the BTC delegation becomes
			active at the current height
		*/
		msgs := h.GenerateCovenantSignaturesMessages(r, covenantSKs, msgCreateBTCDel, actualDel)
		for i := 0; i < int(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum); i++ {
			_, err = h.MsgServer.AddCovenantSigs(h.Ctx, msgs[i])
			h.NoError(err)
		}

		events = h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(h.Ctx, btcTip.Height, btcTip.Height)
		require.Len(t, events, 1)
		btcDelStateUpdate = events[0].GetBtcDelStateUpdate()
		require.NotNil(t, btcDelStateUpdate)
		require.Equal(t, expectedStakingTxHash, btcDelStateUpdate.StakingTxHash)
		require.Equal(t, types.BTCDelegationStatus_ACTIVE, btcDelStateUpdate.NewState)

		// ensure this finality provider has voting power at the current height
		babylonHeight += 1
		h.SetCtxHeight(babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(btcTip).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		h.NoError(err)
		require.Equal(t, uint64(stakingValue), h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fp.BtcPk, babylonHeight))

		// ensure event queue is cleared at BTC tip height
		events = h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(h.Ctx, btcTip.Height, btcTip.Height)
		require.Len(t, events, 0)

		/*
			BTC height reaches end height - w, such that the BTC delegation becomes expired
			ensure the finality provider does not have voting power anymore
		*/
		babylonHeight += 1
		h.SetCtxHeight(babylonHeight)
		h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: unbondedHeight}).AnyTimes()
		err = h.BTCStakingKeeper.BeginBlocker(h.Ctx)
		h.NoError(err)
		require.Zero(t, h.BTCStakingKeeper.GetVotingPower(h.Ctx, *fp.BtcPk, babylonHeight))

		// ensure the unbonded event is processed and cleared
		events = h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(h.Ctx, unbondedHeight, unbondedHeight)
		require.Len(t, events, 0)
	})
}
