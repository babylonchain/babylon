package keeper_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func FuzzChainList(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, _, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times with random chain IDs
		numHeaders := datagen.RandomInt(100)
		expectedChainIDs := []string{}
		for i := uint64(0); i < numHeaders; i++ {
			var chainID string
			// simulate the scenario that some headers belong to the same chain
			if i > 0 && datagen.OneInN(2) {
				chainID = expectedChainIDs[rand.Intn(len(expectedChainIDs))]
			} else {
				chainID = datagen.GenRandomHexStr(30)
				expectedChainIDs = append(expectedChainIDs, chainID)
			}
			header := datagen.GenRandomIBCTMHeader(chainID, 0)
			hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
		}

		// make query to get actual chain IDs
		resp, err := zcKeeper.ChainList(ctx, &zctypes.QueryChainListRequest{})
		require.NoError(t, err)
		actualChainIDs := resp.ChainIds

		// sort them and assert equality
		sort.Strings(expectedChainIDs)
		sort.Strings(actualChainIDs)
		require.Equal(t, len(expectedChainIDs), len(actualChainIDs))
		for i := 0; i < len(expectedChainIDs); i++ {
			require.Equal(t, expectedChainIDs[i], actualChainIDs[i])
		}
	})
}

func FuzzChainInfo(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		numForkHeaders := datagen.RandomInt(10) + 1
		SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, 0, numHeaders, numForkHeaders)

		// check if the chain info of is recorded or not
		resp, err := zcKeeper.ChainInfo(ctx, &zctypes.QueryChainInfoRequest{ChainId: czChain.ChainID})
		require.NoError(t, err)
		chainInfo := resp.ChainInfo
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, numForkHeaders, uint64(len(chainInfo.LatestForks.Headers)))
	})
}

func FuzzHeader(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 2
		numForkHeaders := datagen.RandomInt(10) + 1
		headers, forkHeaders := SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, 0, numHeaders, numForkHeaders)

		// find header at a random height and assert correctness against the expected header
		randomHeight := datagen.RandomInt(int(numHeaders - 1))
		resp, err := zcKeeper.Header(ctx, &zctypes.QueryHeaderRequest{ChainId: czChain.ChainID, Height: randomHeight})
		require.NoError(t, err)
		require.Equal(t, headers[randomHeight].Header.LastCommitHash, resp.Header.Hash)
		require.Len(t, resp.ForkHeaders.Headers, 0)

		// find the last header and fork headers then assert correctness
		resp, err = zcKeeper.Header(ctx, &zctypes.QueryHeaderRequest{ChainId: czChain.ChainID, Height: numHeaders - 1})
		require.NoError(t, err)
		require.Equal(t, headers[numHeaders-1].Header.LastCommitHash, resp.Header.Hash)
		require.Len(t, resp.ForkHeaders.Headers, int(numForkHeaders))
		for i := 0; i < int(numForkHeaders); i++ {
			require.Equal(t, forkHeaders[i].Header.LastCommitHash, resp.ForkHeaders.Headers[i].Hash)
		}
	})
}

func FuzzEpochChainInfo(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		numReqs := datagen.RandomInt(5) + 1

		epochNumList := []uint64{datagen.RandomInt(10) + 1}
		nextHeightList := []uint64{0}
		numHeadersList := []uint64{}
		numForkHeadersList := []uint64{}

		// we test the scenario of ending an epoch for multiple times, in order to ensure that
		// consecutive epoch infos do not affect each other.
		for i := uint64(0); i < numReqs; i++ {
			// generate a random number of headers and fork headers
			numHeadersList = append(numHeadersList, datagen.RandomInt(100)+1)
			numForkHeadersList = append(numForkHeadersList, datagen.RandomInt(10)+1)
			// trigger hooks to append these headers and fork headers
			SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, nextHeightList[i], numHeadersList[i], numForkHeadersList[i])
			// prepare nextHeight for the next request
			nextHeightList = append(nextHeightList, nextHeightList[i]+numHeadersList[i])

			// simulate the scenario that a random epoch has ended
			hooks.AfterEpochEnds(ctx, epochNumList[i])
			// prepare epochNum for the next request
			epochNumList = append(epochNumList, epochNumList[i]+datagen.RandomInt(10)+1)
		}

		// attest the correctness of epoch info for each tested epoch
		for i := uint64(0); i < numReqs; i++ {
			resp, err := zcKeeper.EpochChainInfo(ctx, &zctypes.QueryEpochChainInfoRequest{EpochNum: epochNumList[i], ChainId: czChain.ChainID})
			require.NoError(t, err)
			chainInfo := resp.ChainInfo
			require.Equal(t, nextHeightList[i+1]-1, chainInfo.LatestHeader.Height)
			require.Equal(t, numForkHeadersList[i], uint64(len(chainInfo.LatestForks.Headers)))
		}
	})
}

func FuzzListHeaders(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		numForkHeaders := datagen.RandomInt(10) + 1
		headers, _ := SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, 0, numHeaders, numForkHeaders)

		// a request with randomised pagination
		limit := datagen.RandomInt(int(numHeaders)) + 1
		req := &zctypes.QueryListHeadersRequest{
			ChainId: czChain.ChainID,
			Pagination: &query.PageRequest{
				Limit: limit,
			},
		}
		resp, err := zcKeeper.ListHeaders(ctx, req)
		require.NoError(t, err)
		require.Equal(t, int(limit), len(resp.Headers))
		for i := uint64(0); i < limit; i++ {
			require.Equal(t, headers[i].Header.LastCommitHash, resp.Headers[i].Hash)
		}
	})
}

func FuzzListEpochHeaders(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper
		epochingKeeper := babylonApp.EpochingKeeper

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		numReqs := datagen.RandomInt(5) + 1

		epochNumList := []uint64{datagen.RandomInt(10) + 1}
		nextHeightList := []uint64{0}
		numHeadersList := []uint64{}
		expectedHeadersMap := map[uint64][]*ibctmtypes.Header{}
		numForkHeadersList := []uint64{}

		// we test the scenario of ending an epoch for multiple times, in order to ensure that
		// consecutive epoch infos do not affect each other.
		for i := uint64(0); i < numReqs; i++ {
			epochNum := epochNumList[i]
			// enter a random epoch
			if i == 0 {
				for j := uint64(0); j < epochNum; j++ {
					epochingKeeper.IncEpoch(ctx)
				}
			} else {
				for j := uint64(0); j < epochNum-epochNumList[i-1]; j++ {
					epochingKeeper.IncEpoch(ctx)
				}
			}

			// generate a random number of headers and fork headers
			numHeadersList = append(numHeadersList, datagen.RandomInt(100)+1)
			numForkHeadersList = append(numForkHeadersList, datagen.RandomInt(10)+1)
			// trigger hooks to append these headers and fork headers
			expectedHeaders, _ := SimulateHeadersAndForksViaHook(ctx, hooks, czChain.ChainID, nextHeightList[i], numHeadersList[i], numForkHeadersList[i])
			expectedHeadersMap[epochNum] = expectedHeaders
			// prepare nextHeight for the next request
			nextHeightList = append(nextHeightList, nextHeightList[i]+numHeadersList[i])

			// simulate the scenario that a random epoch has ended
			hooks.AfterEpochEnds(ctx, epochNum)
			// prepare epochNum for the next request
			epochNumList = append(epochNumList, epochNum+datagen.RandomInt(10)+1)
		}

		// attest the correctness of epoch info for each tested epoch
		for i := uint64(0); i < numReqs; i++ {
			epochNum := epochNumList[i]
			// make request
			req := &zctypes.QueryListEpochHeadersRequest{
				ChainId:  czChain.ChainID,
				EpochNum: epochNum,
			}
			resp, err := zcKeeper.ListEpochHeaders(ctx, req)
			require.NoError(t, err)

			// check if the headers are same as expected
			headers := resp.Headers
			require.Equal(t, len(expectedHeadersMap[epochNum]), len(headers))
			for j := 0; j < len(expectedHeadersMap[epochNum]); j++ {
				require.Equal(t, expectedHeadersMap[epochNum][j].Header.LastCommitHash, headers[j].Hash)
			}
		}
	})
}

func FuzzFinalizedChainInfo(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		czChainIDLen := datagen.RandomInt(50) + 1
		czChainID := string(datagen.GenRandomByteArray(czChainIDLen))

		// simulate the scenario that a random epoch has ended and finalised
		epoch := datagen.GenRandomEpoch()

		// mock checkpointing keeper
		// TODO: tests with a set of validators
		checkpointingKeeper := zctypes.NewMockCheckpointingKeeper(ctrl)
		checkpointingKeeper.EXPECT().GetBLSPubKeySet(gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return([]*checkpointingtypes.ValidatorWithBlsKey{}, nil).AnyTimes()
		// mock btccheckpoint keeper
		// TODO: test with BTCSpvProofs
		randomRawCkpt := datagen.GenRandomRawCheckpoint()
		randomRawCkpt.EpochNum = epoch.EpochNumber
		btccKeeper := zctypes.NewMockBtcCheckpointKeeper(ctrl)
		checkpointingKeeper.EXPECT().GetRawCheckpoint(gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return(
			&checkpointingtypes.RawCheckpointWithMeta{
				Ckpt: randomRawCkpt,
			}, nil,
		).AnyTimes()
		btccKeeper.EXPECT().GetFinalizedEpochDataWithBestSubmission(gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return(
			btcctypes.Finalized,
			&btcctypes.SubmissionKey{
				Key: []*btcctypes.TransactionKey{},
			},
			nil,
		).AnyTimes()
		mockSubmissionData := &btcctypes.SubmissionData{TxsInfo: []*btcctypes.TransactionInfo{}}
		btccKeeper.EXPECT().GetSubmissionData(gomock.Any(), gomock.Any()).Return(mockSubmissionData).AnyTimes()
		// mock epoching keeper
		epochingKeeper := zctypes.NewMockEpochingKeeper(ctrl)
		epochingKeeper.EXPECT().GetEpoch(gomock.Any()).Return(epoch).AnyTimes()
		epochingKeeper.EXPECT().GetHistoricalEpoch(gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return(epoch, nil).AnyTimes()
		epochingKeeper.EXPECT().ProveAppHashInEpoch(gomock.Any(), gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return(&tmcrypto.Proof{}, nil).AnyTimes()

		// mock Tendermint client
		// TODO: integration tests with Tendermint
		tmClient := zctypes.NewMockTMClient(ctrl)
		resTx := &tmrpctypes.ResultTx{
			Proof: tmtypes.TxProof{},
		}
		tmClient.EXPECT().Tx(gomock.Any(), gomock.Any(), true).Return(resTx, nil).AnyTimes()

		zcKeeper, ctx := testkeeper.ZoneConciergeKeeper(t, checkpointingKeeper, btccKeeper, epochingKeeper, tmClient)
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100) + 1
		numForkHeaders := datagen.RandomInt(10) + 1
		SimulateHeadersAndForksViaHook(ctx, hooks, czChainID, 0, numHeaders, numForkHeaders)

		hooks.AfterEpochEnds(ctx, epoch.EpochNumber)
		err := hooks.AfterRawCheckpointFinalized(ctx, epoch.EpochNumber)
		require.NoError(t, err)

		// check if the chain info of this epoch is recorded or not
		resp, err := zcKeeper.FinalizedChainInfo(ctx, &zctypes.QueryFinalizedChainInfoRequest{ChainId: czChainID, Prove: true})
		require.NoError(t, err)
		chainInfo := resp.FinalizedChainInfo
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, numForkHeaders, uint64(len(chainInfo.LatestForks.Headers)))
	})
}
