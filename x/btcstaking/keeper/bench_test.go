package keeper_test

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/babylonchain/babylon/testutil/datagen"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	bsmodule "github.com/babylonchain/babylon/x/btcstaking"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/golang/mock/gomock"
)

func benchBeginBlock(b *testing.B, numFPs int, numDelsUnderFP int) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

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
			stakingTxHash, _, _, msgCreateBTCDel, actualDel := h.CreateDelegation(
				r,
				fp.BtcPk.MustToBTCPK(),
				changeAddress.EncodeAddress(),
				stakingValue,
				1000,
			)
			// retrieve BTC delegation in DB
			btcDelMap[stakingTxHash] = append(btcDelMap[stakingTxHash], actualDel)
			// generate and insert new covenant signatures
			// after that, all BTC delegations will have voting power
			h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)
		}
	}

	// mock stuff
	h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()

	// Start the CPU profiler
	cpuProfileFile := fmt.Sprintf("/tmp/btcstaking-beginblock-%d-%d-cpu.pprof", numFPs, numDelsUnderFP)
	f, err := os.Create(cpuProfileFile)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		b.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// Reset timer before the benchmark loop starts
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = bsmodule.BeginBlocker(h.Ctx, *h.BTCStakingKeeper)
		h.NoError(err)
	}
}

func BenchmarkBeginBlock_10_1(b *testing.B)    { benchBeginBlock(b, 10, 1) }
func BenchmarkBeginBlock_10_10(b *testing.B)   { benchBeginBlock(b, 10, 10) }
func BenchmarkBeginBlock_10_100(b *testing.B)  { benchBeginBlock(b, 10, 100) }
func BenchmarkBeginBlock_100_1(b *testing.B)   { benchBeginBlock(b, 100, 1) }
func BenchmarkBeginBlock_100_10(b *testing.B)  { benchBeginBlock(b, 100, 10) }
func BenchmarkBeginBlock_100_100(b *testing.B) { benchBeginBlock(b, 100, 100) }
