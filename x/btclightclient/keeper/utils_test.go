package keeper_test

import (
	"math/big"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Mock hooks interface
// We create a custom type of hooks that just store what they were called with
// to aid with testing.
var _ types.BTCLightClientHooks = &MockHooks{}

type MockHooks struct {
	AfterBTCRollForwardStore    []*types.BTCHeaderInfo
	AfterBTCRollBackStore       []*types.BTCHeaderInfo
	AfterBTCHeaderInsertedStore []*types.BTCHeaderInfo
}

func NewMockHooks() *MockHooks {
	rollForwardStore := make([]*types.BTCHeaderInfo, 0)
	rollBackwardStore := make([]*types.BTCHeaderInfo, 0)
	headerInsertedStore := make([]*types.BTCHeaderInfo, 0)
	return &MockHooks{
		AfterBTCRollForwardStore:    rollForwardStore,
		AfterBTCRollBackStore:       rollBackwardStore,
		AfterBTCHeaderInsertedStore: headerInsertedStore,
	}
}

func (m *MockHooks) AfterBTCRollForward(_ sdk.Context, headerInfo *types.BTCHeaderInfo) {
	m.AfterBTCRollForwardStore = append(m.AfterBTCRollForwardStore, headerInfo)
}

func (m *MockHooks) AfterBTCRollBack(_ sdk.Context, headerInfo *types.BTCHeaderInfo) {
	m.AfterBTCRollBackStore = append(m.AfterBTCRollBackStore, headerInfo)
}

func (m *MockHooks) AfterBTCHeaderInserted(_ sdk.Context, headerInfo *types.BTCHeaderInfo) {
	m.AfterBTCHeaderInsertedStore = append(m.AfterBTCHeaderInsertedStore, headerInfo)
}

func allFieldsEqual(a *types.BTCHeaderInfo, b *types.BTCHeaderInfo) bool {
	return a.Height == b.Height && a.Hash.Eq(b.Hash) && a.Header.Eq(b.Header) && a.Work.Equal(*b.Work)
}

// this function must not be used at difficulty adjustment boundaries, as then
// difficulty adjustment calculation will fail
func genRandomChain(
	t *testing.T,
	r *rand.Rand,
	k *keeper.Keeper,
	ctx sdk.Context,
	initialHeight uint64,
	chainLength uint64,
) (*types.BTCHeaderInfo, *datagen.BTCHeaderPartialChain) {
	genesisHeader := datagen.NewBTCHeaderChainWithLength(r, initialHeight, 0, 1)
	genesisHeaderInfo := genesisHeader.GetChainInfo()[0]
	k.SetBaseBTCHeader(ctx, *genesisHeaderInfo)
	randomChain := datagen.NewBTCHeaderChainFromParentInfo(
		r,
		genesisHeaderInfo,
		uint32(chainLength),
	)
	err := k.InsertHeaders(ctx, randomChain.ChainToBytes())
	require.NoError(t, err)
	tip := k.GetTipInfo(ctx)
	randomChainTipInfo := randomChain.GetTipInfo()
	require.True(t, allFieldsEqual(tip, randomChainTipInfo))
	return genesisHeaderInfo, randomChain
}

func checkTip(
	t *testing.T,
	ctx sdk.Context,
	blcKeeper *keeper.Keeper,
	expectedWork sdkmath.Uint,
	expectedHeight uint64,
	expectedTipHeader *wire.BlockHeader) {

	currentTip := blcKeeper.GetTipInfo(ctx)
	blockByHeight := blcKeeper.GetHeaderByHeight(ctx, currentTip.Height)
	blockByHash := blcKeeper.GetHeaderByHash(ctx, currentTip.Hash)

	// Consistency check between tip and block by height and block by hash
	require.NotNil(t, blockByHeight)
	require.NotNil(t, currentTip)
	require.NotNil(t, blockByHash)
	require.True(t, allFieldsEqual(currentTip, blockByHeight))
	require.True(t, allFieldsEqual(currentTip, blockByHash))

	// check all tip fields
	require.True(t, currentTip.Work.Equal(expectedWork))
	require.Equal(t, currentTip.Height, expectedHeight)
	expectedTipHeaderHash := expectedTipHeader.BlockHash()
	require.True(t, currentTip.Hash.ToChainhash().IsEqual(&expectedTipHeaderHash))
	require.True(t, currentTip.Header.Hash().ToChainhash().IsEqual(&expectedTipHeaderHash))
}

func chainToChainBytes(chain []*wire.BlockHeader) []bbn.BTCHeaderBytes {
	chainBytes := make([]bbn.BTCHeaderBytes, len(chain))
	for i, header := range chain {
		chainBytes[i] = bbn.NewBTCHeaderBytesFromBlockHeader(header)
	}
	return chainBytes
}

func chainWork(chain []*wire.BlockHeader) *sdkmath.Uint {
	totalWork := sdkmath.NewUint(0)
	for _, header := range chain {
		totalWork = sdkmath.NewUintFromBigInt(new(big.Int).Add(totalWork.BigInt(), blockchain.CalcWork(header.Bits)))
	}
	return &totalWork
}
