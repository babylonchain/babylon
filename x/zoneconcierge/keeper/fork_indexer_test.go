package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func FuzzForkIndexer(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, zcKeeper := SetupTest(t)

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100)
		for i := uint64(0); i < numHeaders; i++ {
			header := &ibctmtypes.Header{
				SignedHeader: &tmproto.SignedHeader{
					Header: &tmproto.Header{
						ChainID: czChain.ChainID,
						Height:  int64(i),
					},
				},
			}
			hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
		}

		// generate a number of fork headers
		numForkHeaders := int(datagen.RandomInt(10))
		for i := 0; i < numForkHeaders; i++ {
			header := &ibctmtypes.Header{
				SignedHeader: &tmproto.SignedHeader{
					Header: &tmproto.Header{
						ChainID: czChain.ChainID,
						Height:  int64(numHeaders - 1),
					},
				},
			}
			hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, true)
		}

		// check if the chain info is updated or not
		forks := zcKeeper.GetForks(ctx, czChain.ChainID, numHeaders-1)
		require.Equal(t, numForkHeaders, len(forks.Headers))
	})
}
