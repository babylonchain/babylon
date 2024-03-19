package keeper_test

import (
	"context"
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/wire"
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

func (m *MockHooks) AfterBTCRollForward(_ context.Context, headerInfo *types.BTCHeaderInfo) {
	m.AfterBTCRollForwardStore = append(m.AfterBTCRollForwardStore, headerInfo)
}

func (m *MockHooks) AfterBTCRollBack(_ context.Context, headerInfo *types.BTCHeaderInfo) {
	m.AfterBTCRollBackStore = append(m.AfterBTCRollBackStore, headerInfo)
}

func (m *MockHooks) AfterBTCHeaderInserted(_ context.Context, headerInfo *types.BTCHeaderInfo) {
	m.AfterBTCHeaderInsertedStore = append(m.AfterBTCHeaderInsertedStore, headerInfo)
}

func allFieldsEqual(a *types.BTCHeaderInfo, b *types.BTCHeaderInfo) bool {
	return a.Height == b.Height && a.Hash.Eq(b.Hash) && a.Header.Eq(b.Header) && a.Work.Equal(*b.Work)
}

func checkTip(
	t *testing.T,
	ctx context.Context,
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

func chainWork(chain []*wire.BlockHeader) *sdkmath.Uint {
	totalWork := sdkmath.NewUint(0)
	for _, header := range chain {
		totalWork = sdkmath.NewUintFromBigInt(new(big.Int).Add(totalWork.BigInt(), blockchain.CalcWork(header.Bits)))
	}
	return &totalWork
}
