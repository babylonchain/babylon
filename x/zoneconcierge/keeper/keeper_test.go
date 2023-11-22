package keeper_test

import (
	"context"
	"math/rand"

	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"github.com/babylonchain/babylon/testutil/datagen"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
)

// SimulateNewHeaders generates a non-zero number of canonical headers
func SimulateNewHeaders(ctx context.Context, r *rand.Rand, k *zckeeper.Keeper, chainID string, startHeight uint64, numHeaders uint64) []*ibctmtypes.Header {
	headers := []*ibctmtypes.Header{}
	// invoke the hook a number of times to simulate a number of blocks
	for i := uint64(0); i < numHeaders; i++ {
		header := datagen.GenRandomIBCTMHeader(r, chainID, startHeight+i)
		k.HandleHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(r, 32), datagen.HeaderToHeaderInfo(header), false)
		headers = append(headers, header)
	}
	return headers
}

// SimulateNewHeadersAndForks generates a random non-zero number of canonical headers and fork headers
func SimulateNewHeadersAndForks(ctx context.Context, r *rand.Rand, k *zckeeper.Keeper, chainID string, startHeight uint64, numHeaders uint64, numForkHeaders uint64) ([]*ibctmtypes.Header, []*ibctmtypes.Header) {
	headers := []*ibctmtypes.Header{}
	// invoke the hook a number of times to simulate a number of blocks
	for i := uint64(0); i < numHeaders; i++ {
		header := datagen.GenRandomIBCTMHeader(r, chainID, startHeight+i)
		k.HandleHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(r, 32), datagen.HeaderToHeaderInfo(header), false)
		headers = append(headers, header)
	}

	// generate a number of fork headers
	forkHeaders := []*ibctmtypes.Header{}
	for i := uint64(0); i < numForkHeaders; i++ {
		header := datagen.GenRandomIBCTMHeader(r, chainID, startHeight+numHeaders-1)
		k.HandleHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(r, 32), datagen.HeaderToHeaderInfo(header), true)
		forkHeaders = append(forkHeaders, header)
	}
	return headers, forkHeaders
}
