package babylon_integration

import (
	"context"
	"errors"
	"time"

	tm "github.com/cosmos/cosmos-sdk/client/grpc/tmservice"

	"github.com/babylonchain/babylon/app"
	appparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
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
	signerInfo keyring.Info
	chainId    string
	Conn       *grpc.ClientConn
}

func NewTestTxSender(
	keyringPath string,
	genesisPath string,
	conn *grpc.ClientConn,
) (*TestTxSender, error) {

	kb, err := keyring.New("babylond", "test", keyringPath, nil)

	if err != nil {
		return nil, err
	}

	genDoc, err := types.GenesisDocFromFile(genesisPath)

	if err != nil {
		return nil, err
	}

	infos, _ := kb.List()

	signerInfo := infos[0]

	return &TestTxSender{
		keyring:    kb,
		encConfig:  app.MakeTestEncodingConfig(),
		signerInfo: signerInfo,
		chainId:    genDoc.ChainID,
		Conn:       conn,
	}, nil
}

func (b *TestTxSender) getSenderAddress() ctypes.AccAddress {
	return b.signerInfo.GetAddress()
}

func (b *TestTxSender) buildTx(msg ctypes.Msg, fees string, gas uint64, seqNr uint64) []byte {
	txFactory := tx.Factory{}

	txFactory = txFactory.
		WithKeybase(b.keyring).
		WithTxConfig(b.encConfig.TxConfig).
		WithChainID(b.chainId).
		WithFees(fees).
		WithGas(gas).
		WithSequence(seqNr)

	txb1, _ := txFactory.BuildUnsignedTx(msg)

	if err := tx.Sign(txFactory, b.signerInfo.GetName(), txb1, true); err != nil {
		panic("Tx should sign")
	}

	txBytes, err := b.encConfig.TxConfig.TxEncoder()(txb1.GetTx())

	if err != nil {
		panic("Tx should encode")
	}

	return txBytes
}

func (b *TestTxSender) insertNewEmptyHeader(currentTip *lightclient.BTCHeaderInfo) (*txservice.BroadcastTxResponse, error) {
	childHeaderHex := generateEmptyChildHeaderHexBytes(currentTip.Header.ToBlockHeader())

	address := b.getSenderAddress()

	msg, err := lightclient.NewMsgInsertHeader(address, childHeaderHex)

	if err != nil {
		panic("creating new header message must success ")
	}

	acc, err := b.getAccount()

	if err != nil {
		panic("retrieving sending account must succeed")
	}

	//TODO 3stake and 300000 should probably not be hardcoded by taken from tx
	//simulation. For now this enough to pay for insert header transaction.
	txBytes := b.buildTx(msg, "3stake", 300000, acc.GetSequence())

	req := txservice.BroadcastTxRequest{TxBytes: txBytes, Mode: txservice.BroadcastMode_BROADCAST_MODE_SYNC}

	sender := txservice.NewServiceClient(b.Conn)

	return sender.BroadcastTx(context.Background(), &req)
}

func (b *TestTxSender) getBtcTip() (*lightclient.BTCHeaderInfo, error) {
	lc := lightclient.NewQueryClient(b.Conn)

	res, err := lc.Tip(context.Background(), lightclient.NewQueryTipRequest())

	if err != nil {
		return nil, err
	}

	return res.Header, nil
}

func (b *TestTxSender) getAccount() (acctypes.AccountI, error) {
	queryClient := acctypes.NewQueryClient(b.Conn)

	res, _ := queryClient.Account(
		context.Background(),
		&acctypes.QueryAccountRequest{Address: b.getSenderAddress().String()},
	)

	var acc acctypes.AccountI
	if err := b.encConfig.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return nil, err
	}

	return acc, nil
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

func generateEmptyChildHeaderHexBytes(bh *wire.BlockHeader) string {
	childHeader := generateEmptyChildHeader(bh)
	return bbn.NewBTCHeaderBytesFromBlockHeader(childHeader).MarshalHex()
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
	return WaitForBtcHeightWithTimeout(c, h, 30*time.Second)
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
