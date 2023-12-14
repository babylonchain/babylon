package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testhelper "github.com/babylonchain/babylon/testutil/helper"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
)

func FuzzProofCZHeaderInEpoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		h := testhelper.NewHelper(t)
		ek := h.App.EpochingKeeper
		zck := h.App.ZoneConciergeKeeper
		var err error

		// chain is at height 1 thus epoch 1

		// enter the 1st block of epoch 2
		epochInterval := ek.GetParams(h.Ctx).EpochInterval
		for j := 0; j < int(epochInterval); j++ {
			h.Ctx, err = h.GenAndApplyEmptyBlock(r)
			h.NoError(err)
		}

		// handle a random header from a random consumer chain
		chainID := datagen.GenRandomHexStr(r, 10)
		height := datagen.RandomInt(r, 100) + 1
		ibctmHeader := datagen.GenRandomIBCTMHeader(r, chainID, height)
		headerInfo := datagen.HeaderToHeaderInfo(ibctmHeader)
		zck.HandleHeaderWithValidCommit(h.Ctx, datagen.GenRandomByteArray(r, 32), headerInfo, false)

		// ensure the header is successfully inserted
		indexedHeader, err := zck.GetHeader(h.Ctx, chainID, height)
		h.NoError(err)

		// enter the 1st block of the next epoch
		for j := 0; j < int(epochInterval); j++ {
			h.Ctx, err = h.GenAndApplyEmptyBlock(r)
			h.NoError(err)
		}
		// sealer header hash is the AppHash of the 1st header in CometBFT
		//, i.e., 2nd header captured at BeginBlock() of 2nd header, of an epoch
		sealerHeaderHash := h.Ctx.HeaderInfo().AppHash
		// seal last epoch
		h.Ctx, err = h.GenAndApplyEmptyBlock(r)
		h.NoError(err)

		epochWithHeader, err := ek.GetHistoricalEpoch(h.Ctx, indexedHeader.BabylonEpoch)
		h.NoError(err)
		epochWithHeader.SealerHeaderHash = sealerHeaderHash

		// generate inclusion proof
		proof, err := zck.ProveCZHeaderInEpoch(h.Ctx, indexedHeader, epochWithHeader)
		h.NoError(err)

		// verify the inclusion proof
		err = zckeeper.VerifyCZHeaderInEpoch(indexedHeader, epochWithHeader, proof)
		h.NoError(err)
	})
}
