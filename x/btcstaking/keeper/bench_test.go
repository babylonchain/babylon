package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/testutil/datagen"
	bsmodule "github.com/babylonchain/babylon/x/btcstaking"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/golang/mock/gomock"
)

func BenchmarkBTCStaking_BeginBlock(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	var (
		numFPs         = 10  // 100 finality providers
		numDelsUnderFP = 100 // 100 * 1000 = 100_000 BTC delegations
	)

	// helper
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
	btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
	h := NewHelper(b, btclcKeeper, btccKeeper)
	// set all parameters
	covenantSKs, _ := h.GenAndApplyParams(r)
	changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
	h.NoError(err)

	// generate new finality providers
	fps := []*types.FinalityProvider{}
	for i := 0; i < numFPs; i++ {
		fp, err := datagen.GenRandomFinalityProvider(r)
		h.NoError(err)
		msg := &types.MsgCreateFinalityProvider{
			Signer:      datagen.GenRandomAccount().Address,
			Description: fp.Description,
			Commission:  fp.Commission,
			BabylonPk:   fp.BabylonPk,
			BtcPk:       fp.BtcPk,
			Pop:         fp.Pop,
		}
		_, err = h.MsgServer.CreateFinalityProvider(h.Ctx, msg)
		h.NoError(err)
		fps = append(fps, fp)
	}

	// create new BTC delegations under each finality provider
	btcDelMap := map[string][]*types.BTCDelegation{}
	for _, fp := range fps {
		for i := 0; i < numDelsUnderFP; i++ {
			// generate and insert new BTC delegation
			stakingValue := int64(2 * 10e8)
			stakingTxHash, _, _, msgCreateBTCDel := h.CreateDelegation(
				r,
				fp.BtcPk.MustToBTCPK(),
				changeAddress.EncodeAddress(),
				stakingValue,
				1000,
			)
			// retrieve BTC delegation in DB
			actualDel, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
			h.NoError(err)
			btcDelMap[stakingTxHash] = append(btcDelMap[stakingTxHash], actualDel)
			// generate and insert new covenant signatures
			// after that, all BTC delegations will have voting power
			h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)
		}
	}

	// Reset timer before the benchmark loop starts
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = bsmodule.BeginBlocker(h.Ctx, *h.BTCStakingKeeper)
		h.NoError(err)
	}
}
