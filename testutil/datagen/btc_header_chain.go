package datagen

import (
	"math/rand"

	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

// Init header is always simnet header due to need to solve pow
var initHeader = chaincfg.SimNetParams.GenesisBlock.Header

type BTCHeaderPartialChain struct {
	// slice of Headers forming valid chain
	Headers                  []*wire.BlockHeader
	initialHeaderHeight      uint64
	inititialHeaderTotalWork sdkmath.Uint
}

func NewBTCHeaderChainWithLength(
	r *rand.Rand,
	initialHeaderHeight uint64,
	initialHeaderTotalWork uint64,
	length uint32) *BTCHeaderPartialChain {
	return NewBTCHeaderChainFromParent(
		r,
		initialHeaderHeight,
		sdkmath.NewUint(initialHeaderTotalWork),
		&initHeader,
		length,
	)
}

func NewBTCHeaderChainFromParentInfo(
	r *rand.Rand,
	parent *types.BTCHeaderInfo,
	length uint32,
) *BTCHeaderPartialChain {
	return NewBTCHeaderChainFromParent(
		r,
		parent.Height+1,
		*parent.Work,
		parent.Header.ToBlockHeader(),
		length,
	)
}

func NewBTCHeaderChainFromParent(
	r *rand.Rand,
	initialHeaderHeight uint64,
	initialHeaderTotalWork sdkmath.Uint,
	parent *wire.BlockHeader,
	length uint32,
) *BTCHeaderPartialChain {
	headers := GenRandomValidChainStartingFrom(
		r,
		initialHeaderHeight,
		parent,
		nil,
		length,
	)
	return &BTCHeaderPartialChain{
		Headers:                  headers,
		initialHeaderHeight:      initialHeaderHeight,
		inititialHeaderTotalWork: initialHeaderTotalWork,
	}
}

func (c *BTCHeaderPartialChain) GetChainInfo() []*types.BTCHeaderInfo {
	return ChainToInfoChain(c.Headers, c.initialHeaderHeight, c.inititialHeaderTotalWork)
}

func (c *BTCHeaderPartialChain) GetChainInfoResponse() []*types.BTCHeaderInfoResponse {
	return ChainToInfoResponseChain(c.Headers, c.initialHeaderHeight, c.inititialHeaderTotalWork)
}

func (c *BTCHeaderPartialChain) ChainToBytes() []bbn.BTCHeaderBytes {
	chainBytes := make([]bbn.BTCHeaderBytes, 0)
	for _, header := range c.Headers {
		h := header
		bytes := bbn.NewBTCHeaderBytesFromBlockHeader(h)
		chainBytes = append(chainBytes, bytes)
	}

	return chainBytes
}

func (c *BTCHeaderPartialChain) GetTipInfo() *types.BTCHeaderInfo {
	chainInfo := ChainToInfoChain(c.Headers, c.initialHeaderHeight, c.inititialHeaderTotalWork)
	return chainInfo[len(chainInfo)-1]
}

func (c *BTCHeaderPartialChain) TipHeader() *wire.BlockHeader {
	return c.Headers[len(c.Headers)-1]
}

func (c *BTCHeaderPartialChain) GetRandomHeaderInfo(r *rand.Rand) *types.BTCHeaderInfo {
	randIdx := RandomInt(r, len(c.Headers))
	headerInfo := ChainToInfoChain(c.Headers, c.initialHeaderHeight, c.inititialHeaderTotalWork)
	return headerInfo[randIdx]
}

func (c *BTCHeaderPartialChain) GetRandomHeaderInfoNoTip(r *rand.Rand) *types.BTCHeaderInfo {
	randIdx := RandomInt(r, len(c.Headers)-1)
	headerInfo := ChainToInfoChain(c.Headers, c.initialHeaderHeight, c.inititialHeaderTotalWork)
	return headerInfo[randIdx]
}

func (c *BTCHeaderPartialChain) ChainLength() int {
	return len(c.Headers)
}

func (c *BTCHeaderPartialChain) GetHeadersMap() map[string]*wire.BlockHeader {
	headersMap := make(map[string]*wire.BlockHeader)
	for _, header := range c.Headers {
		headersMap[header.BlockHash().String()] = header
	}
	return headersMap
}
