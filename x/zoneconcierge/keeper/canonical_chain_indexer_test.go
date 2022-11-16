package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func FuzzCanonicalChainIndexer(f *testing.F) {
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
						ChainID:        czChain.ChainID,
						Height:         int64(i),
						LastCommitHash: datagen.GenRandomByteArray(32),
					},
				},
			}
			hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
		}

		// check if the canonical chain index is correct or not
		for i := uint64(0); i < numHeaders; i++ {
			header, err := zcKeeper.GetHeader(ctx, czChain.ChainID, i)
			require.NoError(t, err)
			require.NotNil(t, header)

			require.Equal(t, czChain.ChainID, header.ChainId)
			require.Equal(t, i, header.Height)
		}
	})
}
