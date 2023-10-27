package types

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type BtcChainReadStore interface {
	GetHeaderByHash(hash *bbn.BTCHeaderHashBytes) (*BTCHeaderInfo, error)
	GetHeaderByHeight(height uint64) (*BTCHeaderInfo, error)
	GetTip() *BTCHeaderInfo
}

// Copy from neutrino light client
// https://github.com/lightninglabs/neutrino/blob/master/blockmanager.go#L2875
type lightChainCtx struct {
	params              *chaincfg.Params
	blocksPerRetarget   int32
	minRetargetTimespan int64
	maxRetargetTimespan int64
}

var _ blockchain.ChainCtx = (*lightChainCtx)(nil)

func newLightChainCtx(params *chaincfg.Params, blocksPerRetarget int32,
	minRetargetTimespan, maxRetargetTimespan int64) *lightChainCtx {

	return &lightChainCtx{
		params:              params,
		blocksPerRetarget:   blocksPerRetarget,
		minRetargetTimespan: minRetargetTimespan,
		maxRetargetTimespan: maxRetargetTimespan,
	}
}

func newLightChainCtxFromParams(params *chaincfg.Params) *lightChainCtx {
	targetTimespan := int64(params.TargetTimespan / time.Second)
	targetTimePerBlock := int64(params.TargetTimePerBlock / time.Second)
	adjustmentFactor := params.RetargetAdjustmentFactor

	blocksPerRetarget := int32(targetTimespan / targetTimePerBlock)
	minRetargetTimespan := targetTimespan / adjustmentFactor
	maxRetargetTimespan := targetTimespan * adjustmentFactor

	return newLightChainCtx(
		params, blocksPerRetarget, minRetargetTimespan, maxRetargetTimespan,
	)
}

func (l *lightChainCtx) ChainParams() *chaincfg.Params {
	return l.params
}

func (l *lightChainCtx) BlocksPerRetarget() int32 {
	return l.blocksPerRetarget
}

func (l *lightChainCtx) MinRetargetTimespan() int64 {
	return l.minRetargetTimespan
}

func (l *lightChainCtx) MaxRetargetTimespan() int64 {
	return l.maxRetargetTimespan
}

// We never check checkpoints in our on-chain light client. Required by blockchain.ChainCtx interface
func (l *lightChainCtx) VerifyCheckpoint(int32, *chainhash.Hash) bool {
	return false
}

// If VerifyCheckpoint returns false, this function is never called. Required by blockchain.ChainCtx interface
func (l *lightChainCtx) FindPreviousCheckpoint() (blockchain.HeaderCtx, error) {
	return nil, nil
}

type localHeaderInfo struct {
	header    *wire.BlockHeader
	height    uint64
	totalWork sdkmath.Uint
}

func newLocalHeaderInfo(
	header *wire.BlockHeader,
	height uint64,
	totalWork sdkmath.Uint) *localHeaderInfo {

	return &localHeaderInfo{
		header:    header,
		height:    height,
		totalWork: totalWork,
	}
}

func toLocalInfo(i *BTCHeaderInfo) *localHeaderInfo {
	if i == nil {
		return nil
	}

	return newLocalHeaderInfo(i.Header.ToBlockHeader(), i.Height, *i.Work)
}

func (i *localHeaderInfo) toBTCHeaderInfo() *BTCHeaderInfo {
	blockHash := i.header.BlockHash()
	headerBytes := bbn.NewBTCHeaderBytesFromBlockHeader(i.header)
	headerHash := bbn.NewBTCHeaderHashBytesFromChainhash(&blockHash)

	return NewBTCHeaderInfo(
		&headerBytes,
		&headerHash,
		i.height,
		&i.totalWork,
	)
}

func toBTCHeaderInfos(infos []*localHeaderInfo) []*BTCHeaderInfo {
	result := make([]*BTCHeaderInfo, len(infos))

	for i, info := range infos {
		result[i] = info.toBTCHeaderInfo()
	}

	return result
}

// based on neutrio light client
// https://github.com/lightninglabs/neutrino/blob/master/blockmanager.go#L2944
type lightHeaderCtx struct {
	height    uint64
	bits      uint32
	timestamp int64
	store     *storeWithExtensionChain
}

var _ blockchain.HeaderCtx = (*lightHeaderCtx)(nil)

func newLightHeaderCtx(height uint64, header *wire.BlockHeader,
	store *storeWithExtensionChain) *lightHeaderCtx {

	return &lightHeaderCtx{
		height:    height,
		bits:      header.Bits,
		timestamp: header.Timestamp.Unix(),
		store:     store,
	}
}

func (l *lightHeaderCtx) Height() int32 {
	return int32(l.height)
}

func (l *lightHeaderCtx) Bits() uint32 {
	return l.bits
}

func (l *lightHeaderCtx) Timestamp() int64 {
	return l.timestamp
}

func (l *lightHeaderCtx) Parent() blockchain.HeaderCtx {
	// The parent is just an ancestor with distance 1.
	anc := l.RelativeAncestorCtx(1)

	if anc == nil {
		return nil
	}

	return anc
}

func (l *lightHeaderCtx) RelativeAncestorCtx(
	distance int32) blockchain.HeaderCtx {

	ancestorHeight := l.Height() - distance

	if ancestorHeight < 0 {
		// We don't have this header.
		return nil
	}

	ancU64 := uint64(ancestorHeight)

	ancestor := l.store.getHeaderAtHeight(ancU64)

	if ancestor == nil {
		return nil
	}

	return newLightHeaderCtx(
		ancU64, ancestor.header, l.store,
	)
}

type BtcLightClient struct {
	params *chaincfg.Params
	ctx    *lightChainCtx
}

func NewBtcLightClient(
	params *chaincfg.Params,
	ctx *lightChainCtx) *BtcLightClient {
	return &BtcLightClient{
		params: params,
		ctx:    ctx,
	}
}

func NewBtcLightClientFromParams(params *chaincfg.Params) *BtcLightClient {
	return NewBtcLightClient(params, newLightChainCtxFromParams(params))
}

func headersFormChain(headers []*wire.BlockHeader) bool {
	var (
		lastHeader chainhash.Hash
		emptyHash  chainhash.Hash
	)
	for _, blockHeader := range headers {
		blockHash := blockHeader.BlockHash()

		// If we haven't yet set lastHeader, set it now.
		if lastHeader == emptyHash {
			lastHeader = blockHash
			continue
		}

		// Ensure that blockHeader.PrevBlock matches lastHeader.
		if blockHeader.PrevBlock != lastHeader {
			return false
		}

		lastHeader = blockHash
	}

	return true
}

type DisableHeaderInTheFutureValidationTimeSource struct {
	h *wire.BlockHeader
}

func NewDisableHeaderInTheFutureValidationTimeSource(header *wire.BlockHeader) *DisableHeaderInTheFutureValidationTimeSource {
	return &DisableHeaderInTheFutureValidationTimeSource{
		h: header,
	}
}

// AdjustedTime returns the current time adjusted by the median time
// offset as calculated from the time samples added by AddTimeSample.
func (d *DisableHeaderInTheFutureValidationTimeSource) AdjustedTime() time.Time {
	return d.h.Timestamp
}

func (d *DisableHeaderInTheFutureValidationTimeSource) AddTimeSample(id string, timeVal time.Time) {
	//no op
}

func (d *DisableHeaderInTheFutureValidationTimeSource) Offset() time.Duration {
	return 0 * time.Second
}

type storeWithExtensionChain struct {
	headers []*localHeaderInfo
	store   BtcChainReadStore
}

func newStoreWithExtensionChain(
	store BtcChainReadStore,
	maxExentsionHeaders int,
) *storeWithExtensionChain {

	return &storeWithExtensionChain{
		// large capacity to avoid reallocation
		headers: make([]*localHeaderInfo, 0, maxExentsionHeaders),
		store:   store,
	}
}

func (s *storeWithExtensionChain) addHeader(header *localHeaderInfo) {
	s.headers = append(s.headers, header)
}

func (s *storeWithExtensionChain) getHeaderAtHeight(height uint64) *localHeaderInfo {
	if len(s.headers) == 0 || height < s.headers[0].height {
		h, err := s.store.GetHeaderByHeight(height)

		if err != nil {
			return nil
		}
		return newLocalHeaderInfo(h.Header.ToBlockHeader(), height, *h.Work)
	} else {
		headerIndex := height - s.headers[0].height
		return s.headers[headerIndex]
	}
}

func (l *BtcLightClient) processNewHeadersChain(
	store *storeWithExtensionChain,
	chainParent *localHeaderInfo,
	chain []*wire.BlockHeader) error {
	// init info about parent as current tip
	parentHeaderInfo := chainParent

	for _, blockHeader := range chain {
		h := blockHeader

		err := l.checkHeader(
			store, parentHeaderInfo, h,
		)

		if err != nil {
			return err
		}

		childWork := CalcHeaderWork(h)
		newHeaderInfo := newLocalHeaderInfo(
			h,
			parentHeaderInfo.height+1,
			CumulativeWork(parentHeaderInfo.totalWork, childWork),
		)
		store.addHeader(newHeaderInfo)
		parentHeaderInfo = newHeaderInfo
	}

	return nil
}

type RollbackInfo struct {
	HeaderToRollbackTo *BTCHeaderInfo
}

type InsertResult struct {
	HeadersToInsert []*BTCHeaderInfo
	// if rollback is not nil, it means that we need to rollback to the header provided header
	RollbackInfo *RollbackInfo
}

func (l *BtcLightClient) InsertHeaders(readStore BtcChainReadStore, headers []*wire.BlockHeader) (*InsertResult, error) {
	headersLen := len(headers)
	if headersLen == 0 {
		return nil, fmt.Errorf("cannot insert empty headers")
	}

	if !headersFormChain(headers) {
		return nil, fmt.Errorf("headers do not form a chain")
	}

	currentTip := toLocalInfo(readStore.GetTip())

	if currentTip == nil {
		return nil, fmt.Errorf("cannot insert headers when tip is nil")
	}

	currentTipHash := currentTip.header.BlockHash()

	firstHeaderOfExtensionChain := headers[0]

	store := newStoreWithExtensionChain(readStore, headersLen)

	if firstHeaderOfExtensionChain.PrevBlock.IsEqual(&currentTipHash) {
		// most common case - extending of current tip
		if err := l.processNewHeadersChain(store, currentTip, headers); err != nil {
			return nil, err
		}

		return &InsertResult{
			HeadersToInsert: toBTCHeaderInfos(store.headers),
			RollbackInfo:    nil,
		}, nil
	} else {
		// here we received potential new fork
		parentHash := bbn.NewBTCHeaderHashBytesFromChainhash(&firstHeaderOfExtensionChain.PrevBlock)
		forkParent, err := readStore.GetHeaderByHash(&parentHash)

		if err != nil {
			return nil, err
		}

		forkParentInfo := toLocalInfo(forkParent)

		if err := l.processNewHeadersChain(store, forkParentInfo, headers); err != nil {
			return nil, err
		}

		tipOfNewChain := store.headers[len(store.headers)-1]

		if tipOfNewChain.totalWork.LTE(currentTip.totalWork) {
			return nil, fmt.Errorf("the new chain has less or equal work than the current tip")
		}

		return &InsertResult{
			HeadersToInsert: toBTCHeaderInfos(store.headers),
			RollbackInfo: &RollbackInfo{
				// we need to rollback to fork parent
				HeaderToRollbackTo: forkParent,
			},
		}, nil
	}
}

// checkHeader checks if the header is valid and can be added to the store.
// One criticial condition is that to properly validate difficulty adjustments
// we should have at least one header which is at difficulty adjustment boundary
// in store.
func (l *BtcLightClient) checkHeader(
	s *storeWithExtensionChain,
	parentHeaderInfo *localHeaderInfo,
	blockHeader *wire.BlockHeader,
) error {
	parentHeaderCtx := newLightHeaderCtx(
		parentHeaderInfo.height, parentHeaderInfo.header, s,
	)

	var emptyFlags blockchain.BehaviorFlags
	err := blockchain.CheckBlockHeaderContext(
		blockHeader, parentHeaderCtx, emptyFlags, l.ctx, true,
	)
	if err != nil {
		return err
	}

	return blockchain.CheckBlockHeaderSanity(
		blockHeader, l.params.PowLimit, NewDisableHeaderInTheFutureValidationTimeSource(blockHeader),
		emptyFlags,
	)
}
