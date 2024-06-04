package datagen

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
	btclightclientk "github.com/babylonchain/babylon/x/btclightclient/keeper"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

type RetargetInfo struct {
	LastRetargetHeader *wire.BlockHeader
	Params             *chaincfg.Params
}

type TimeBetweenBlocksInfo struct {
	Time time.Duration
}

// Difficulty calculation copied from btcd
// https://github.com/btcsuite/btcd/blob/master/blockchain/difficulty.go#L221
func calculateAdjustedDifficulty(
	lastRetargetHeader *wire.BlockHeader,
	currentHeaderTimestamp time.Time,
	params *chaincfg.Params) uint32 {

	targetTimespan := int64(params.TargetTimespan / time.Second)
	adjustmentFactor := params.RetargetAdjustmentFactor
	minRetargetTimespan := targetTimespan / adjustmentFactor
	maxRetargetTimespan := targetTimespan * adjustmentFactor

	// Limit the amount of adjustment that can occur to the previous
	// difficulty.
	actualTimespan := currentHeaderTimestamp.Unix() - lastRetargetHeader.Timestamp.Unix()
	adjustedTimespan := actualTimespan
	if actualTimespan < minRetargetTimespan {
		adjustedTimespan = minRetargetTimespan
	} else if actualTimespan > maxRetargetTimespan {
		adjustedTimespan = maxRetargetTimespan
	}

	// Calculate new target difficulty as:
	//  currentDifficulty * (adjustedTimespan / targetTimespan)
	// The result uses integer division which means it will be slightly
	// rounded down.  Bitcoind also uses integer division to calculate this
	// result.
	oldTarget := blockchain.CompactToBig(lastRetargetHeader.Bits)
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
	newTarget.Div(newTarget, big.NewInt(targetTimespan))

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(params.PowLimit) > 0 {
		newTarget.Set(params.PowLimit)
	}

	newTargetBits := blockchain.BigToCompact(newTarget)
	return newTargetBits
}

func GenRandomBtcdHeader(r *rand.Rand) *wire.BlockHeader {
	version := GenRandomBTCHeaderVersion(r)
	bits := GenRandomBTCHeaderBits(r)
	nonce := GenRandomBTCHeaderNonce(r)
	prevBlock := GenRandomBTCHeaderPrevBlock(r)
	merkleRoot := GenRandomBTCHeaderMerkleRoot(r)
	timestamp := GenRandomBTCHeaderTimestamp(r)

	header := &wire.BlockHeader{
		Version:    version,
		Bits:       bits,
		Nonce:      nonce,
		PrevBlock:  *prevBlock.ToChainhash(),
		MerkleRoot: *merkleRoot,
		Timestamp:  timestamp,
	}
	return header
}

// GenRandomBTCHeaderBits constructs a random uint32 corresponding to BTC header difficulty bits
func GenRandomBTCHeaderBits(r *rand.Rand) uint32 {
	// Instead of navigating through all the different signs and bit constructing
	// of the workBits, we can resort to having a uint64 (instead of the maximum of 2^256).
	// First, generate an integer, convert it into a big.Int and then into compact form.

	difficulty := r.Uint64()
	if difficulty == 0 {
		difficulty += 1
	}
	bigDifficulty := sdkmath.NewUint(difficulty)

	workBits := blockchain.BigToCompact(bigDifficulty.BigInt())
	return workBits
}

// GenRandomBTCHeaderPrevBlock constructs a random BTCHeaderHashBytes instance
func GenRandomBTCHeaderPrevBlock(r *rand.Rand) *bbn.BTCHeaderHashBytes {
	hex := GenRandomHexStr(r, bbn.BTCHeaderHashLen)
	hashBytes, _ := bbn.NewBTCHeaderHashBytesFromHex(hex)
	return &hashBytes
}

// GenRandomBTCHeaderMerkleRoot generates a random hex string corresponding to a merkle root
func GenRandomBTCHeaderMerkleRoot(r *rand.Rand) *chainhash.Hash {
	// TODO: the length should become a constant and the merkle root should have a custom type
	chHash, _ := chainhash.NewHashFromStr(GenRandomHexStr(r, 32))
	return chHash
}

// GenRandomBTCHeaderTimestamp generates a random BTC header timestamp
func GenRandomBTCHeaderTimestamp(r *rand.Rand) time.Time {
	randomTime := r.Int63n(time.Now().Unix())
	return time.Unix(randomTime, 0)
}

// GenRandomBTCHeaderVersion generates a random version integer
func GenRandomBTCHeaderVersion(r *rand.Rand) int32 {
	return r.Int31()
}

// GenRandomBTCHeaderNonce generates a random BTC header nonce
func GenRandomBTCHeaderNonce(r *rand.Rand) uint32 {
	return r.Uint32()
}

// GenRandomBTCHeaderBytes generates a random BTCHeaderBytes object
// based on randomly generated BTC header attributes
// If the `parent` argument is not `nil`, then the `PrevBlock`
// attribute of the BTC header will point to the hash of the parent and the
// `Timestamp` attribute will be later than the parent's `Timestamp`.
// If the `bitsBig` argument is not `nil`, then the `Bits` attribute
// of the BTC header will point to the compact form of big integer.
func GenRandomBTCHeaderBytes(r *rand.Rand, parent *btclightclienttypes.BTCHeaderInfo, bitsBig *sdkmath.Uint) bbn.BTCHeaderBytes {
	btcdHeader := GenRandomBtcdHeader(r)

	if bitsBig != nil {
		btcdHeader.Bits = blockchain.BigToCompact(bitsBig.BigInt())
	}
	if parent != nil {
		// Set the parent hash
		btcdHeader.PrevBlock = *parent.Hash.ToChainhash()
		// The time should be more recent than the parent time
		// Typical BTC header difference is 10 mins with some fluctuations
		// The header timestamp is going to be the time of the parent + 10 mins +- 0-59 seconds
		seconds := r.Intn(60)
		if OneInN(r, 2) { // 50% of the times subtract the seconds
			seconds = -1 * seconds
		}
		btcdHeader.Timestamp = parent.Header.Time().Add(time.Minute*10 + time.Duration(seconds)*time.Second)
	}

	return bbn.NewBTCHeaderBytesFromBlockHeader(btcdHeader)
}

// GenRandomBTCHeight returns a random uint64
func GenRandomBTCHeight(r *rand.Rand) uint64 {
	return r.Uint64()
}

// GenRandomBTCHeaderInfoWithParentAndBits generates a BTCHeaderInfo object in which the `header.PrevBlock` points to the `parent`
// and the `Work` property points to the accumulated work (parent.Work + header.Work). Less bits as a parameter, means more difficulty.
func GenRandomBTCHeaderInfoWithParentAndBits(r *rand.Rand, parent *btclightclienttypes.BTCHeaderInfo, bits *sdkmath.Uint) *btclightclienttypes.BTCHeaderInfo {
	header := GenRandomBTCHeaderBytes(r, parent, bits)
	height := GenRandomBTCHeight(r)
	if parent != nil {
		height = parent.Height + 1
	}

	accumulatedWork := btclightclienttypes.CalcWork(&header)
	if parent != nil {
		accumulatedWork = btclightclienttypes.CumulativeWork(accumulatedWork, *parent.Work)
	}

	return &btclightclienttypes.BTCHeaderInfo{
		Header: &header,
		Hash:   header.Hash(),
		Height: height,
		Work:   &accumulatedWork,
	}
}

// GenRandomBTCHeaderInfoWithParent generates a random BTCHeaderInfo object
// in which the parent points to the `parent` parameter.
func GenRandomBTCHeaderInfoWithParent(r *rand.Rand, parent *btclightclienttypes.BTCHeaderInfo) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomBTCHeaderInfoWithParentAndBits(r, parent, nil)
}

// GenRandomValidBTCHeaderInfoWithParent generates random BTCHeaderInfo object
// with valid proof of work.
// WARNING: if parent is from network with a lot of work (mainnet) it may never finish
// use only with simnet headers
func GenRandomValidBTCHeaderInfoWithParent(r *rand.Rand, parent btclightclienttypes.BTCHeaderInfo) *btclightclienttypes.BTCHeaderInfo {
	parentHeader := parent.Header.ToBlockHeader()
	randHeader := GenRandomBtcdValidHeader(r, parentHeader, nil, nil)
	headerBytes := bbn.NewBTCHeaderBytesFromBlockHeader(randHeader)

	accumulatedWork := btclightclienttypes.CalcWork(&headerBytes)
	accumulatedWork = btclightclienttypes.CumulativeWork(accumulatedWork, *parent.Work)

	return &btclightclienttypes.BTCHeaderInfo{
		Header: &headerBytes,
		Hash:   headerBytes.Hash(),
		Height: parent.Height + 1,
		Work:   &accumulatedWork,
	}
}

// random duration between 4 and 12mins in seconds
func GenRandomTimeBetweenBlocks(r *rand.Rand) time.Duration {
	return time.Duration(r.Int63n(8*60)+4*60) * time.Second
}

func GenRandomBtcdValidHeader(
	r *rand.Rand,
	parent *wire.BlockHeader,
	timeAfterParent *TimeBetweenBlocksInfo,
	retargetInfo *RetargetInfo,
) *wire.BlockHeader {
	randHeader := GenRandomBtcdHeader(r)
	randHeader.Version = 4
	randHeader.PrevBlock = parent.BlockHash()

	if timeAfterParent == nil {
		// random time after
		randHeader.Timestamp = parent.Timestamp.Add(GenRandomTimeBetweenBlocks(r))
	} else {
		randHeader.Timestamp = parent.Timestamp.Add(timeAfterParent.Time)
	}

	if retargetInfo == nil {
		// If no retarget info is provided, then we assume that the difficulty is the same as the parent
		randHeader.Bits = parent.Bits
	} else {
		// If retarget info is provided, then we calculate the difficulty based on the info provided
		randHeader.Bits = calculateAdjustedDifficulty(
			retargetInfo.LastRetargetHeader,
			parent.Timestamp,
			retargetInfo.Params,
		)
	}
	SolveBlock(randHeader)
	return randHeader
}

func GenRandomValidChainStartingFrom(
	r *rand.Rand,
	parentHeaderHeight uint64,
	parentHeader *wire.BlockHeader,
	timeBetweenBlocks *TimeBetweenBlocksInfo,
	numHeaders uint32,
) []*wire.BlockHeader {
	if numHeaders == 0 {
		return []*wire.BlockHeader{}
	}

	headers := make([]*wire.BlockHeader, numHeaders)
	for i := uint32(0); i < numHeaders; i++ {
		if i == 0 {
			headers[i] = GenRandomBtcdValidHeader(r, parentHeader, timeBetweenBlocks, nil)
			continue
		}

		headers[i] = GenRandomBtcdValidHeader(r, headers[i-1], timeBetweenBlocks, nil)
	}
	return headers
}

// GenRandBtcChainInsertingInKeeper generates random BTCHeaderInfo and insert its headers
// into the keeper store.
// this function must not be used at difficulty adjustment boundaries, as then
// difficulty adjustment calculation will fail
func GenRandBtcChainInsertingInKeeper(
	t *testing.T,
	r *rand.Rand,
	k *btclightclientk.Keeper,
	ctx context.Context,
	initialHeight uint64,
	chainLength uint64,
) (*btclightclienttypes.BTCHeaderInfo, *BTCHeaderPartialChain) {
	genesisHeader := NewBTCHeaderChainWithLength(r, initialHeight, 0, 1)
	genesisHeaderInfo := genesisHeader.GetChainInfo()[0]
	k.SetBaseBTCHeader(ctx, *genesisHeaderInfo)
	randomChain := NewBTCHeaderChainFromParentInfo(
		r,
		genesisHeaderInfo,
		uint32(chainLength),
	)
	err := k.InsertHeaders(ctx, randomChain.ChainToBytes())
	require.NoError(t, err)
	tip := k.GetTipInfo(ctx)
	randomChainTipInfo := randomChain.GetTipInfo()
	require.True(t, tip.Eq(randomChainTipInfo))
	return genesisHeaderInfo, randomChain
}

func ChainToInfoChain(
	chain []*wire.BlockHeader,
	initialHeaderNumber uint64,
	initialHeaderTotalWork sdkmath.Uint,
) []*btclightclienttypes.BTCHeaderInfo {
	if len(chain) == 0 {
		return []*btclightclienttypes.BTCHeaderInfo{}
	}

	infoChain := make([]*btclightclienttypes.BTCHeaderInfo, len(chain))

	totalDifficulty := initialHeaderTotalWork

	for i, header := range chain {
		headerWork := btclightclienttypes.CalcHeaderWork(header)
		headerTotalDifficulty := btclightclienttypes.CumulativeWork(headerWork, totalDifficulty)
		hash := header.BlockHash()
		headerBytes := bbn.NewBTCHeaderBytesFromBlockHeader(header)
		headerHash := bbn.NewBTCHeaderHashBytesFromChainhash(&hash)
		headerNumber := initialHeaderNumber + uint64(i)

		headerInfo := btclightclienttypes.NewBTCHeaderInfo(
			&headerBytes,
			&headerHash,
			headerNumber,
			&headerTotalDifficulty,
		)

		infoChain[i] = headerInfo

		totalDifficulty = headerTotalDifficulty
	}

	return infoChain
}

func ChainToInfoResponseChain(
	chain []*wire.BlockHeader,
	initialHeaderNumber uint64,
	initialHeaderTotalWork sdkmath.Uint,
) []*btclightclienttypes.BTCHeaderInfoResponse {
	if len(chain) == 0 {
		return []*btclightclienttypes.BTCHeaderInfoResponse{}
	}

	infoChain := make([]*btclightclienttypes.BTCHeaderInfoResponse, len(chain))

	totalDifficulty := initialHeaderTotalWork

	for i, header := range chain {
		headerWork := btclightclienttypes.CalcHeaderWork(header)
		headerTotalDifficulty := btclightclienttypes.CumulativeWork(headerWork, totalDifficulty)
		hash := header.BlockHash()
		headerBytes := bbn.NewBTCHeaderBytesFromBlockHeader(header)
		headerHash := bbn.NewBTCHeaderHashBytesFromChainhash(&hash)
		headerNumber := initialHeaderNumber + uint64(i)

		headerInfoResponse := btclightclienttypes.NewBTCHeaderInfoResponse(
			&headerBytes,
			&headerHash,
			headerNumber,
			&headerTotalDifficulty,
		)

		infoChain[i] = headerInfoResponse

		totalDifficulty = headerTotalDifficulty
	}

	return infoChain
}

func HeaderToHeaderBytes(headers []*wire.BlockHeader) []bbn.BTCHeaderBytes {
	headerBytes := make([]bbn.BTCHeaderBytes, len(headers))
	for i, header := range headers {
		headerBytes[i] = bbn.NewBTCHeaderBytesFromBlockHeader(header)
	}
	return headerBytes
}

func GenRandomBTCHeaderInfoWithBits(r *rand.Rand, bits *sdkmath.Uint) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomBTCHeaderInfoWithParentAndBits(r, nil, bits)
}

// GenRandomBTCHeaderInfo generates a random BTCHeaderInfo object
func GenRandomBTCHeaderInfo(r *rand.Rand) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomBTCHeaderInfoWithParent(r, nil)
}

func GenRandomBTCHeaderInfoWithInvalidHeader(r *rand.Rand, powLimit *big.Int) *btclightclienttypes.BTCHeaderInfo {
	var tries = 0
	for {
		info := GenRandomBTCHeaderInfo(r)

		err := bbn.ValidateBTCHeader(info.Header.ToBlockHeader(), powLimit)

		if err != nil {
			return info
		}

		tries++

		if tries >= 100 {
			panic("Failed to generate invalid btc header in 100 random tries")
		}
	}
}

// MutateHash takes a hash as a parameter, copies it, modifies the copy, and returns the copy.
func MutateHash(r *rand.Rand, hash *bbn.BTCHeaderHashBytes) *bbn.BTCHeaderHashBytes {
	mutatedBytes := make([]byte, bbn.BTCHeaderHashLen)
	// Retrieve a random byte index
	idx := RandomInt(r, bbn.BTCHeaderHashLen)
	copy(mutatedBytes, hash.MustMarshal())
	// Add one to the index
	mutatedBytes[idx] += 1
	mutated, _ := bbn.NewBTCHeaderHashBytesFromBytes(mutatedBytes)
	return &mutated
}
