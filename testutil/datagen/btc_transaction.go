package datagen

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/rand"
	"runtime"
	"time"

	txformat "github.com/babylonchain/babylon/btctxformatter"

	"github.com/babylonchain/babylon/btctxformatter"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// opTrueScript is simply a public key script that contains the OP_TRUE
	// opcode.  It is defined here to reduce garbage creation.
	opTrueScript = []byte{txscript.OP_TRUE}

	tranasctionVersion = 1

	lowFee = btcutil.Amount(1)
)

type testCheckpointData struct {
	epoch            uint64
	lastCommitHash   []byte
	bitmap           []byte
	blsSig           []byte
	submitterAddress []byte
}

type TestRawCheckpointData struct {
	Epoch            uint64
	FirstPart        []byte
	SecondPart       []byte
	ExpectedOpReturn []byte
}

// standardCoinbaseScript returns a standard script suitable for use as the
// signature script of the coinbase transaction of a new block.  In particular,
// it starts with the block height that is required by version 2 blocks
func standardCoinbaseScript(blockHeight int32, extraNonce uint64) ([]byte, error) {
	return txscript.NewScriptBuilder().AddInt64(int64(blockHeight)).
		AddInt64(int64(extraNonce)).Script()
}

// opReturnScript returns a provably-pruneable OP_RETURN script with the
// provided data.(i.e  this is how OP_RETURN script should look like to not pollue
// unspent transaction outputs )
func opReturnScript(data []byte) []byte {
	builder := txscript.NewScriptBuilder()
	script, err := builder.AddOp(txscript.OP_RETURN).AddData(data).Script()
	if err != nil {
		panic(err)
	}
	return script
}

// SolveBlock mutates provided header.Nonce, until it solves proof of work puzzle
// WARNING: providing header with too large difficulty (header.Bits) will make this
// function unfinishable.
func SolveBlock(header *wire.BlockHeader) bool {
	// sbResult is used by the solver goroutines to send results.
	type sbResult struct {
		found bool
		nonce uint32
	}

	// solver accepts a block header and a nonce range to test. It is
	// intended to be run as a goroutine.
	targetDifficulty := blockchain.CompactToBig(header.Bits)
	quit := make(chan bool)
	results := make(chan sbResult)
	solver := func(hdr wire.BlockHeader, startNonce, stopNonce uint32) {
		// We need to modify the nonce field of the header, so make sure
		// we work with a copy of the original header.
		for i := startNonce; i >= startNonce && i <= stopNonce; i++ {
			select {
			case <-quit:
				return
			default:
				hdr.Nonce = i
				hash := hdr.BlockHash()
				if blockchain.HashToBig(&hash).Cmp(
					targetDifficulty) <= 0 {

					results <- sbResult{true, i}
					return
				}
			}
		}
		results <- sbResult{false, 0}
	}

	startNonce := uint32(1)
	stopNonce := uint32(math.MaxUint32)
	numCores := uint32(runtime.NumCPU())
	noncesPerCore := (stopNonce - startNonce) / numCores
	for i := uint32(0); i < numCores; i++ {
		rangeStart := startNonce + (noncesPerCore * i)
		rangeStop := startNonce + (noncesPerCore * (i + 1)) - 1
		if i == numCores-1 {
			rangeStop = stopNonce
		}
		go solver(*header, rangeStart, rangeStop)
	}
	for i := uint32(0); i < numCores; i++ {
		result := <-results
		if result.found {
			close(quit)
			header.Nonce = result.nonce
			return true
		}
	}

	return false
}

func calcMerkleRoot(txns []*wire.MsgTx) chainhash.Hash {
	if len(txns) == 0 {
		return chainhash.Hash{}
	}

	utilTxns := make([]*btcutil.Tx, 0, len(txns))
	for _, tx := range txns {
		utilTxns = append(utilTxns, btcutil.NewTx(tx))
	}
	merkles := blockchain.BuildMerkleTreeStore(utilTxns, false)
	return *merkles[len(merkles)-1]
}

func createCoinbaseTx(blockHeight int32, params *chaincfg.Params) *wire.MsgTx {
	extraNonce := uint64(0)
	coinbaseScript, err := standardCoinbaseScript(blockHeight, extraNonce)
	if err != nil {
		panic(err)
	}

	tx := wire.NewMsgTx(int32(tranasctionVersion))
	tx.AddTxIn(&wire.TxIn{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOutPoint: *wire.NewOutPoint(&chainhash.Hash{},
			wire.MaxPrevOutIndex),
		Sequence:        wire.MaxTxInSequenceNum,
		SignatureScript: coinbaseScript,
	})
	tx.AddTxOut(&wire.TxOut{
		Value:    blockchain.CalcBlockSubsidy(blockHeight, params),
		PkScript: opTrueScript,
	})
	return tx
}

func uniqueOpReturnScript(r *rand.Rand) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data[0:8], r.Uint64())
	return opReturnScript(data)
}

type spendableOut struct {
	prevOut wire.OutPoint
	amount  btcutil.Amount
}

func randOutPoint(r *rand.Rand) wire.OutPoint {
	hash, _ := chainhash.NewHash(GenRandomByteArray(r, chainhash.HashSize))
	// TODO this will be deterministic without seed but for now it is not that
	// important
	idx := r.Uint32()

	return wire.OutPoint{
		Hash:  *hash,
		Index: idx,
	}
}

func makeSpendableOutWithRandOutPoint(r *rand.Rand, amount btcutil.Amount) spendableOut {
	out := randOutPoint(r)

	return spendableOut{
		prevOut: out,
		amount:  amount,
	}
}

func createSpendTx(r *rand.Rand, spend *spendableOut, fee btcutil.Amount) *wire.MsgTx {
	spendTx := wire.NewMsgTx(int32(tranasctionVersion))
	spendTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: spend.prevOut,
		Sequence:         wire.MaxTxInSequenceNum,
		SignatureScript:  nil,
	})
	spendTx.AddTxOut(wire.NewTxOut(int64(spend.amount-fee),
		opTrueScript))
	// uniqueOpReturnScript is needed so that each transactions is different have
	// different hash
	spendTx.AddTxOut(wire.NewTxOut(0, uniqueOpReturnScript(r)))

	return spendTx
}

func createSpendOpReturnTx(spend *spendableOut, fee btcutil.Amount, data []byte) *wire.MsgTx {
	spendTx := wire.NewMsgTx(int32(tranasctionVersion))
	spendTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: spend.prevOut,
		Sequence:         wire.MaxTxInSequenceNum,
		SignatureScript:  nil,
	})
	spendTx.AddTxOut(wire.NewTxOut(int64(spend.amount-fee),
		opTrueScript))
	spendTx.AddTxOut(wire.NewTxOut(0, opReturnScript(data)))

	return spendTx
}

func CreatOpReturnTransaction(r *rand.Rand, babylonData []byte) *wire.MsgTx {
	out := makeSpendableOutWithRandOutPoint(r, 1000)
	tx := createSpendOpReturnTx(&out, lowFee, babylonData)
	return tx
}

type BlockCreationResult struct {
	HeaderBytes  bbn.BTCHeaderBytes
	Transactions []string
	BbnTxIndex   uint32
}

func CreateBlock(
	r *rand.Rand,
	height uint32,
	numTx uint32,
	babylonOpReturnIdx uint32,
	babylonData []byte,
) *BlockCreationResult {

	if babylonOpReturnIdx > numTx {
		panic("babylon tx index should be less than number of transasactions and greater than 0")
	}

	var transactions []*wire.MsgTx

	for i := uint32(0); i <= numTx; i++ {
		if i == 0 {
			tx := createCoinbaseTx(int32(height), &chaincfg.SimNetParams)
			transactions = append(transactions, tx)
		} else if i == babylonOpReturnIdx {
			out := makeSpendableOutWithRandOutPoint(r, 1000)
			tx := createSpendOpReturnTx(&out, lowFee, babylonData)
			transactions = append(transactions, tx)
		} else {
			out := makeSpendableOutWithRandOutPoint(r, 1000)
			tx := createSpendTx(r, &out, lowFee)
			transactions = append(transactions, tx)
		}
	}

	btcHeader := GenRandomBtcdHeader(r)

	// setting SimNetParams so that block can be easily solved
	btcHeader.Bits = chaincfg.SimNetParams.GenesisBlock.Header.Bits
	btcHeader.MerkleRoot = calcMerkleRoot(transactions)

	solved := SolveBlock(btcHeader)

	if !solved {
		panic("Should solve block")
	}

	var hexTx []string
	for _, tx := range transactions {
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		_ = tx.Serialize(buf)
		hexTx = append(hexTx, hex.EncodeToString(buf.Bytes()))
	}

	res := BlockCreationResult{
		HeaderBytes:  bbn.NewBTCHeaderBytesFromBlockHeader(btcHeader),
		Transactions: hexTx,
		BbnTxIndex:   babylonOpReturnIdx,
	}

	return &res
}

type BtcHeaderWithProof struct {
	HeaderBytes bbn.BTCHeaderBytes
	SpvProof    *btcctypes.BTCSpvProof
}

func CreateBlockWithTransaction(
	r *rand.Rand,
	ph *wire.BlockHeader,
	tx *wire.MsgTx,
) *BtcHeaderWithProof {

	var transactions []*wire.MsgTx
	// height does not matter here, as it is used only for calculation of reward
	transactions = append(transactions, createCoinbaseTx(int32(889), &chaincfg.SimNetParams))
	transactions = append(transactions, tx)

	randHeader := GenRandomBtcdHeader(r)
	randHeader.Version = ph.Version
	randHeader.PrevBlock = ph.BlockHash()
	randHeader.Bits = ph.Bits
	randHeader.Timestamp = ph.Timestamp.Add(50 * time.Second)
	randHeader.MerkleRoot = calcMerkleRoot(transactions)
	SolveBlock(randHeader)

	var txBytes [][]byte
	for _, tx := range transactions {
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		_ = tx.Serialize(buf)
		txBytes = append(txBytes, buf.Bytes())
	}

	headerBytes := bbn.NewBTCHeaderBytesFromBlockHeader(randHeader)

	proof, err := btcctypes.SpvProofFromHeaderAndTransactions(&headerBytes, txBytes, 1)

	if err != nil {
		panic("could not calculate proof")
	}

	return &BtcHeaderWithProof{
		HeaderBytes: headerBytes,
		SpvProof:    proof,
	}
}

func GenRandomTx(r *rand.Rand) *wire.MsgTx {
	// structure of the below tx is from https://github.com/btcsuite/btcd/blob/master/wire/msgtx_test.go
	tx := &wire.MsgTx{
		Version: 1,
		TxIn: []*wire.TxIn{
			{
				PreviousOutPoint: wire.OutPoint{
					Hash:  GenRandomBtcdHash(r),
					Index: r.Uint32(),
				},
				SignatureScript: GenRandomByteArray(r, 10),
				Sequence:        r.Uint32(),
			},
		},
		TxOut: []*wire.TxOut{
			{
				Value:    r.Int63(),
				PkScript: GenRandomByteArray(r, 80),
			},
		},
		LockTime: 0,
	}

	return tx
}

func GenRandomBabylonTxPair(r *rand.Rand) ([]*wire.MsgTx, *btctxformatter.RawBtcCheckpoint) {
	txs := []*wire.MsgTx{GenRandomTx(r), GenRandomTx(r)}
	builder := txscript.NewScriptBuilder()

	// fake a raw checkpoint
	rawBTCCkpt := GetRandomRawBtcCheckpoint(r)
	tag := GenRandomByteArray(r, 4)
	// encode raw checkpoint to two halves
	firstHalf, secondHalf, err := btctxformatter.EncodeCheckpointData(
		btctxformatter.BabylonTag(tag),
		btctxformatter.CurrentVersion,
		rawBTCCkpt,
	)
	if err != nil {
		panic(err)
	}

	dataScript1, err := builder.AddOp(txscript.OP_RETURN).AddData(firstHalf).Script()
	if err != nil {
		panic(err)
	}
	txs[0].TxOut[0] = wire.NewTxOut(0, dataScript1)

	// reset builder
	builder = txscript.NewScriptBuilder()

	dataScript2, err := builder.AddOp(txscript.OP_RETURN).AddData(secondHalf).Script()
	if err != nil {
		panic(err)
	}
	txs[1].TxOut[0] = wire.NewTxOut(0, dataScript2)

	return txs, rawBTCCkpt
}

func GenRandomBabylonTx(r *rand.Rand) *wire.MsgTx {
	txs, _ := GenRandomBabylonTxPair(r)
	idx := r.Intn(2)
	return txs[idx]
}

func BlockCreationResultToProofs(inputs []*BlockCreationResult) []*btcctypes.BTCSpvProof {
	var spvs []*btcctypes.BTCSpvProof

	for _, input := range inputs {
		headerBytes, err := bbn.NewBTCHeaderBytesFromBytes(input.HeaderBytes)

		if err != nil {
			panic("created BlockCreationResult should always contain valid block ")
		}

		var txBytes [][]byte

		for _, t := range input.Transactions {
			tbytes, err := hex.DecodeString(t)

			if err != nil {
				panic("Inputs should contain valid hex encoded transactions")
			}

			txBytes = append(txBytes, tbytes)
		}

		spv, err := btcctypes.SpvProofFromHeaderAndTransactions(&headerBytes, txBytes, uint(input.BbnTxIndex))

		if err != nil {
			panic("Inputs should contain valid spv hex encoded data")
		}

		spvs = append(spvs, spv)
	}

	return spvs
}

func GenerateMessageWithRandomSubmitter(blockResults []*BlockCreationResult) *btcctypes.MsgInsertBTCSpvProof {
	proofs := BlockCreationResultToProofs(blockResults)

	pk, _ := NewPV().GetPubKey()

	address := sdk.AccAddress(pk.Address().Bytes())

	msg := btcctypes.MsgInsertBTCSpvProof{
		Proofs:    proofs,
		Submitter: address.String(),
	}

	return &msg
}

func getRandomCheckpointDataForEpoch(r *rand.Rand, e uint64) testCheckpointData {
	return testCheckpointData{
		epoch:            e,
		lastCommitHash:   GenRandomByteArray(r, txformat.LastCommitHashLength),
		bitmap:           GenRandomByteArray(r, txformat.BitMapLength),
		blsSig:           GenRandomByteArray(r, txformat.BlsSigLength),
		submitterAddress: GenRandomByteArray(r, txformat.AddressLength),
	}
}

// both f and s must be parts retrived from txformat.Encode
func getExpectedOpReturn(tag txformat.BabylonTag, f []byte, s []byte) []byte {
	firstPartNoHeader, err := txformat.GetCheckpointData(
		tag,
		txformat.CurrentVersion,
		0,
		f,
	)

	if err != nil {
		panic("ExpectedOpReturn provided first part should be valid checkpoint data")
	}

	secondPartNoHeader, err := txformat.GetCheckpointData(
		tag,
		txformat.CurrentVersion,
		1,
		s,
	)

	if err != nil {
		panic("ExpectedOpReturn provided second part should be valid checkpoint data")
	}

	connected, err := txformat.ConnectParts(txformat.CurrentVersion, firstPartNoHeader, secondPartNoHeader)

	if err != nil {
		panic("ExpectedOpReturn parts should be connected")
	}

	return connected
}

func RandomRawCheckpointDataForEpoch(r *rand.Rand, e uint64) (*TestRawCheckpointData, *txformat.RawBtcCheckpoint) {
	checkpointData := getRandomCheckpointDataForEpoch(r, e)
	rawBTCCkpt := &txformat.RawBtcCheckpoint{
		Epoch:            checkpointData.epoch,
		LastCommitHash:   checkpointData.lastCommitHash,
		BitMap:           checkpointData.bitmap,
		SubmitterAddress: checkpointData.submitterAddress,
		BlsSig:           checkpointData.blsSig,
	}
	return EncodeRawCkptToTestData(rawBTCCkpt), rawBTCCkpt
}

func EncodeRawCkptToTestData(rawBTCCkpt *txformat.RawBtcCheckpoint) *TestRawCheckpointData {
	tag := btcctypes.DefaultCheckpointTag
	tagAsBytes, _ := hex.DecodeString(tag)
	babylonTag := txformat.BabylonTag(tagAsBytes)
	data1, data2 := txformat.MustEncodeCheckpointData(
		babylonTag,
		txformat.CurrentVersion,
		rawBTCCkpt,
	)
	opReturn := getExpectedOpReturn(babylonTag, data1, data2)

	return &TestRawCheckpointData{
		Epoch:            rawBTCCkpt.Epoch,
		FirstPart:        data1,
		SecondPart:       data2,
		ExpectedOpReturn: opReturn,
	}
}

// test helper to generate random number int in range, min and max must be non-negative
func numInRange(r *rand.Rand, min int, max int) int {
	if min < 0 || max < 0 || min >= max {
		panic("min and max maust be positve numbers. min must be smaller than max")
	}

	return r.Intn(max-min) + min
}

func GenerateMessageWithRandomSubmitterForEpoch(r *rand.Rand, epoch uint64) *btcctypes.MsgInsertBTCSpvProof {
	numTransactions := 100
	tx1 := numInRange(r, 1, 99)

	tx2 := numInRange(r, 1, 99)
	// in those tests epoch is not important
	raw, _ := RandomRawCheckpointDataForEpoch(r, epoch)

	blck1 := CreateBlock(r, 0, uint32(numTransactions), uint32(tx1), raw.FirstPart)

	blck2 := CreateBlock(r, 0, uint32(numTransactions), uint32(tx2), raw.SecondPart)

	msg := GenerateMessageWithRandomSubmitter([]*BlockCreationResult{blck1, blck2})

	return msg
}
