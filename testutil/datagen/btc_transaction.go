package datagen

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/rand"
	"runtime"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var (
	// opTrueScript is simply a public key script that contains the OP_TRUE
	// opcode.  It is defined here to reduce garbage creation.
	opTrueScript = []byte{txscript.OP_TRUE}

	lowFee = btcutil.Amount(1)
)

func standardCoinbaseScript(blockHeight int32, extraNonce uint64) ([]byte, error) {
	return txscript.NewScriptBuilder().AddInt64(int64(blockHeight)).
		AddInt64(int64(extraNonce)).Script()
}

// opReturnScript returns a provably-pruneable OP_RETURN script with the
// provided data.
func opReturnScript(data []byte) []byte {
	builder := txscript.NewScriptBuilder()
	script, err := builder.AddOp(txscript.OP_RETURN).AddData(data).Script()
	if err != nil {
		panic(err)
	}
	return script
}

func solveBlock(header *wire.BlockHeader) bool {
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

	tx := wire.NewMsgTx(1)
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

func uniqueOpReturnScript() []byte {
	rand, err := wire.RandomUint64()
	if err != nil {
		panic(err)
	}

	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data[0:8], rand)
	return opReturnScript(data)
}

type spendableOut struct {
	prevOut wire.OutPoint
	amount  btcutil.Amount
}

func randOutPoint() wire.OutPoint {
	hash, _ := chainhash.NewHash(GenRandomByteArray(chainhash.HashSize))
	// TODO this will be deterministic without seed but for now it is not that
	// important
	idx := rand.Uint32()

	return wire.OutPoint{
		Hash:  *hash,
		Index: idx,
	}
}

func makeSpendableOutWithRandOutPoint(amount btcutil.Amount) spendableOut {
	out := randOutPoint()

	return spendableOut{
		prevOut: out,
		amount:  amount,
	}
}

func createSpendTx(spend *spendableOut, fee btcutil.Amount) *wire.MsgTx {
	spendTx := wire.NewMsgTx(1)
	spendTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: spend.prevOut,
		Sequence:         wire.MaxTxInSequenceNum,
		SignatureScript:  nil,
	})
	spendTx.AddTxOut(wire.NewTxOut(int64(spend.amount-fee),
		opTrueScript))
	// uniqueOpReturnScript is needed so that each transactions is different have
	// different hash
	spendTx.AddTxOut(wire.NewTxOut(0, uniqueOpReturnScript()))

	return spendTx
}

func createSpendOpReturnTx(spend *spendableOut, fee btcutil.Amount, data []byte) *wire.MsgTx {
	spendTx := wire.NewMsgTx(1)
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

type BlockCreationResult struct {
	HeaderBytes  bbn.BTCHeaderBytes
	Transactions []string
	BbnTxIndex   uint32
}

func CreateBlock(
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
			out := makeSpendableOutWithRandOutPoint(1000)
			tx := createSpendOpReturnTx(&out, lowFee, babylonData)
			transactions = append(transactions, tx)
		} else {
			out := makeSpendableOutWithRandOutPoint(1000)
			tx := createSpendTx(&out, lowFee)
			transactions = append(transactions, tx)
		}
	}

	btcHeader := GenRandomBtcdHeader()

	// setting SimNetParams so that block can be easily solved
	btcHeader.Bits = chaincfg.SimNetParams.GenesisBlock.Header.Bits
	btcHeader.MerkleRoot = calcMerkleRoot(transactions)

	solved := solveBlock(btcHeader)

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
