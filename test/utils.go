package babylon_integration

import (
	"context"
	"errors"
	"fmt"
	"time"

	tm "github.com/cosmos/cosmos-sdk/client/grpc/tmservice"

	"github.com/babylonchain/babylon/app"
	appparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	btccheckpoint "github.com/babylonchain/babylon/x/btccheckpoint/types"
	lightclient "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	ctypes "github.com/cosmos/cosmos-sdk/types"
	txservice "github.com/cosmos/cosmos-sdk/types/tx"
	acctypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
)

type TestTxSender struct {
	keyring    keyring.Keyring
	encConfig  appparams.EncodingConfig
	signerInfo *keyring.Record
	chainId    string
	Conn       *grpc.ClientConn
}

func NewTestTxSender(
	keyringPath string,
	genesisPath string,
	conn *grpc.ClientConn,
) (*TestTxSender, error) {
	cfg := app.MakeTestEncodingConfig()

	kb, err := keyring.New("babylond", "test", keyringPath, nil, cfg.Marshaler)

	if err != nil {
		return nil, err
	}

	genDoc, err := types.GenesisDocFromFile(genesisPath)

	if err != nil {
		return nil, err
	}

	signer, err := kb.Key("test-spending-key")

	if err != nil {
		panic("test-spending-key should be defined for each node in integration test")
	}

	signerInfo := signer

	return &TestTxSender{
		keyring:    kb,
		encConfig:  cfg,
		signerInfo: signerInfo,
		chainId:    genDoc.ChainID,
		Conn:       conn,
	}, nil
}

func (b *TestTxSender) getSenderAddress() ctypes.AccAddress {
	addr, err := b.signerInfo.GetAddress()

	if err != nil {
		panic("Getting address from sender should always succeed")
	}
	return addr
}

func (b *TestTxSender) buildTx(fees string, gas uint64, seqNr uint64, accNumber uint64, msgs ...ctypes.Msg) []byte {
	txFactory := tx.Factory{}

	txFactory = txFactory.
		WithKeybase(b.keyring).
		WithTxConfig(b.encConfig.TxConfig).
		WithChainID(b.chainId).
		WithFees(fees).
		WithGas(gas).
		WithSequence(seqNr).
		WithAccountNumber(accNumber)

	txb1, _ := txFactory.BuildUnsignedTx(msgs...)

	if err := tx.Sign(txFactory, b.signerInfo.Name, txb1, true); err != nil {
		panic("Tx should sign")
	}

	txBytes, err := b.encConfig.TxConfig.TxEncoder()(txb1.GetTx())

	if err != nil {
		panic("Tx should encode")
	}

	return txBytes
}

func (b *TestTxSender) SendBtcHeadersTransaction(headers []bbn.BTCHeaderBytes) (*txservice.BroadcastTxResponse, error) {
	if len(headers) == 0 {
		return nil, errors.New("headers should not be empty")
	}

	acc, err := b.getSelfAccount()

	if err != nil {
		panic("retrieving sending account must succeed")
	}

	address := b.getSenderAddress()

	var msgs []ctypes.Msg
	var fees uint64
	var gas uint64

	for _, header := range headers {
		msg, err := lightclient.NewMsgInsertHeader(address, header.MarshalHex())

		if err != nil {
			panic("creating new header message must succeed ")
		}

		msgs = append(msgs, msg)

		fees = fees + 3
		gas = gas + 300000
	}

	feesString := fmt.Sprintf("%d%s", fees, appparams.DefaultBondDenom)

	txBytes := b.buildTx(feesString, gas, acc.GetSequence(), acc.GetAccountNumber(), msgs...)

	req := txservice.BroadcastTxRequest{TxBytes: txBytes, Mode: txservice.BroadcastMode_BROADCAST_MODE_SYNC}

	sender := txservice.NewServiceClient(b.Conn)

	return sender.BroadcastTx(context.Background(), &req)
}

func GenerateNEmptyHeaders(tip *bbn.BTCHeaderBytes, n uint64) []bbn.BTCHeaderBytes {
	var headers []bbn.BTCHeaderBytes

	if n == 0 {
		return headers
	}

	for i := uint64(0); i < n; i++ {
		if i == 0 {
			// first new header, need to use tip as base
			headers = append(headers, generateEmptyChildHeaderBytes(tip))
		} else {
			headers = append(headers, generateEmptyChildHeaderBytes(&headers[i-1]))
		}
	}

	return headers
}

func (b *TestTxSender) insertSpvProof(p1 *btccheckpoint.BTCSpvProof, p2 *btccheckpoint.BTCSpvProof) (*txservice.BroadcastTxResponse, error) {
	address := b.getSenderAddress()

	msg := btccheckpoint.MsgInsertBTCSpvProof{
		Submitter: address.String(),
		Proofs:    []*btccheckpoint.BTCSpvProof{p1, p2},
	}

	acc, err := b.getSelfAccount()

	if err != nil {
		panic("retrieving sending account must succeed")
	}

	fee := fmt.Sprintf("3000%s", appparams.BaseCoinUnit)
	txBytes := b.buildTx(fee, 300000, acc.GetSequence(), acc.GetAccountNumber(), &msg)

	req := txservice.BroadcastTxRequest{TxBytes: txBytes, Mode: txservice.BroadcastMode_BROADCAST_MODE_SYNC}

	sender := txservice.NewServiceClient(b.Conn)

	return sender.BroadcastTx(context.Background(), &req)
}

func (b *TestTxSender) GetBtcTip() *lightclient.BTCHeaderInfo {
	lc := lightclient.NewQueryClient(b.Conn)

	res, err := lc.Tip(context.Background(), lightclient.NewQueryTipRequest())

	if err != nil {
		panic("should retrieve btc header")
	}

	return res.Header
}

func (b *TestTxSender) getAccount(addr ctypes.AccAddress) (acctypes.AccountI, error) {
	queryClient := acctypes.NewQueryClient(b.Conn)

	res, _ := queryClient.Account(
		context.Background(),
		&acctypes.QueryAccountRequest{Address: addr.String()},
	)

	var acc acctypes.AccountI
	if err := b.encConfig.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (b *TestTxSender) getSelfAccount() (acctypes.AccountI, error) {
	return b.getAccount(b.getSenderAddress())
}

func (b *TestTxSender) insertBTCHeaders(currentTip uint64, headers []bbn.BTCHeaderBytes) error {
	lenHeaders := len(headers)

	if lenHeaders == 0 {
		return nil
	}

	_, err := b.SendBtcHeadersTransaction(headers)

	if err != nil {
		return err
	}

	_, err = WaitBtcForHeight(b.Conn, currentTip+uint64(lenHeaders))

	if err != nil {
		return err
	}

	return nil
}

func (b *TestTxSender) insertNEmptyBTCHeaders(n uint64) error {
	currentTip := b.GetBtcTip()
	headers := GenerateNEmptyHeaders(currentTip.Header, n)

	err := b.insertBTCHeaders(currentTip.Height, headers)

	if err != nil {
		return err
	}

	return nil
}

func generateEmptyChildHeader(bh *wire.BlockHeader) *wire.BlockHeader {
	randHeader := datagen.GenRandomBtcdHeader()

	randHeader.Version = bh.Version
	randHeader.PrevBlock = bh.BlockHash()
	randHeader.Bits = bh.Bits
	randHeader.Timestamp = bh.Timestamp.Add(50 * time.Second)
	datagen.SolveBlock(randHeader)

	return randHeader
}

func generateEmptyChildHeaderBytes(bh *bbn.BTCHeaderBytes) bbn.BTCHeaderBytes {
	childHeader := generateEmptyChildHeader(bh.ToBlockHeader())
	return bbn.NewBTCHeaderBytesFromBlockHeader(childHeader)
}

// TODO following helpers could probably be generalized by taking function
// as param
// Tendermint blockchain helpers

func LatestHeight(c *grpc.ClientConn) (int64, error) {
	latestResponse, err := tm.NewServiceClient(c).GetLatestBlock(context.Background(), &tm.GetLatestBlockRequest{})
	if err != nil {
		return 0, err
	}

	return latestResponse.Block.Header.Height, nil
}

func WaitForHeight(c *grpc.ClientConn, h int64) (int64, error) {
	return WaitForHeightWithTimeout(c, h, 15*time.Second)
}

func WaitForHeightWithTimeout(c *grpc.ClientConn, h int64, t time.Duration) (int64, error) {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(t)

	var latestHeight int64

	for {
		select {
		case <-timeout:
			ticker.Stop()
			return latestHeight, errors.New("timeout exceeded waiting for block")
		case <-ticker.C:
			latestH, err := LatestHeight(c)
			if err == nil {
				latestHeight = latestH
				if latestHeight >= h {
					return latestHeight, nil
				}
			}
		}
	}
}

func WaitForNextBlock(c *grpc.ClientConn) error {
	lastBlock, err := LatestHeight(c)
	if err != nil {
		return err
	}

	_, err = WaitForHeight(c, lastBlock+1)

	if err != nil {
		return err
	}

	return nil
}

// Btc blockchain helpers

func BtcLatestHeight(c *grpc.ClientConn) (uint64, error) {
	latestResponse, err := lightclient.NewQueryClient(c).Tip(context.Background(), lightclient.NewQueryTipRequest())
	if err != nil {
		return 0, err
	}

	return latestResponse.Header.Height, nil
}

func WaitForBtcHeightWithTimeout(c *grpc.ClientConn, h uint64, t time.Duration) (uint64, error) {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(t)
	var latestHeight uint64

	for {
		select {
		case <-timeout:
			ticker.Stop()
			return latestHeight, errors.New("timeout exceeded waiting for btc block")
		case <-ticker.C:
			latestH, err := BtcLatestHeight(c)
			if err == nil {
				latestHeight = latestH
				if latestHeight >= h {
					return latestHeight, nil
				}
			}
		}
	}
}

func WaitBtcForHeight(c *grpc.ClientConn, h uint64) (uint64, error) {
	return WaitForBtcHeightWithTimeout(c, h, 15*time.Second)
}

func WaitForNextBtcBlock(c *grpc.ClientConn) error {
	lastBlock, err := BtcLatestHeight(c)
	if err != nil {
		return err
	}

	_, err = WaitBtcForHeight(c, lastBlock+1)

	if err != nil {
		return err
	}

	return nil
}
