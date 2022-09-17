//go:build integration
// +build integration

package babylon_integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	appparams "github.com/babylonchain/babylon/app/params"
	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	lightclient "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	ref "github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	tm "github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"google.golang.org/grpc"
)

// Addresses of all nodes in local testnet.
// TODO: instead of hardcoding them it would be nice to get them from env variables
// so docker-compose and this file would stay compatible
var addresses = []string{
	"localhost:9090",
	"localhost:9091",
	"localhost:9092",
	"localhost:9093",
}

var clients []*grpc.ClientConn

func checkInterfacesWithRetries(client *grpc.ClientConn, maxTries int, sleepTime time.Duration) error {
	tries := 0
	refClient := ref.NewReflectionServiceClient(client)
	for {
		_, err := refClient.ListAllInterfaces(context.Background(), &ref.ListAllInterfacesRequest{})

		if err == nil {
			// successful call to client, finish calling
			return nil
		}

		tries++

		if tries > maxTries {
			return errors.New("Failed to call client")
		}

		<-time.After(sleepTime)
	}
}

func allClientsOverBlockNumber(clients []*grpc.ClientConn, blockNumber int64) bool {
	for _, c := range clients {
		latestResponse, err := tm.NewServiceClient(c).GetLatestBlock(context.Background(), &tm.GetLatestBlockRequest{})

		if err != nil {
			errorString := fmt.Sprintf("Integration tests failed, due to node failure. Erro: %v", err)
			panic(errorString)
		}

		if latestResponse.Block.Header.Height < blockNumber {
			return false
		}
	}
	// we iterated over all clients, and all of them were  >= blockNumber
	return true
}

func waitForBlock(clients []*grpc.ClientConn, blockNumber int64) {
	for {
		allOver := allClientsOverBlockNumber(clients, blockNumber)

		if allOver {
			return
		}

		<-time.After(2 * time.Second)
	}
}

func TestMain(m *testing.M) {

	// This is needed so that all address prefixes are in Babylon format
	appparams.SetAddressPrefixes()

	for _, addr := range addresses {
		grpcConn, err := grpc.Dial(
			addr,                // Or your gRPC server address.
			grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
		)

		if err != nil {
			panic("Grpc connection failed cannot perform integration tests")
		}

		clients = append(clients, grpcConn)
	}
	//runs all following tests
	exitVal := m.Run()

	for _, c := range clients {
		// close all connections after the tests
		c.Close()
	}

	os.Exit(exitVal)
}

// This test serves as a waiting point for testnet to start, it is needed as
// docker compose is usually started in detached mode in CI, therefore tests
// are started even before all nodes are up.
// TODO: investigate starting testnet from golang test file.
func TestTestnetRuninng(t *testing.T) {

	for _, c := range clients {
		err := checkInterfacesWithRetries(c, 40, 5*time.Second)

		if err != nil {
			panic("Could not start integration tests. Testnet not running")
		}
	}
}

// Check all nodes are properly initialized to genesis
// TODO ultimatly we would like to check genesis related to all modules here.
func TestBtcLightClientGenesis(t *testing.T) {
	// The default testnet directory uses the simnet genesis header as its base
	// with height 0.
	hardcodedHeader, _ := bbn.NewBTCHeaderBytesFromHex("0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a45068653ffff7f2002000000")
	hardcodedHeaderHeight := uint64(0)

	for i, c := range clients {
		lc := lightclient.NewQueryClient(c)

		res, err := lc.Tip(context.Background(), lightclient.NewQueryTipRequest())

		if err != nil {
			// this is fatal, as it means we most probably did not get any response and
			// at least one of the nodes is down
			t.Fatalf("Test failed due to client error: %v to node with address %s", err, addresses[i])
		}

		if res.Header.Height != hardcodedHeaderHeight || !res.Header.Hash.Eq(hardcodedHeader.Hash()) {
			t.Errorf("Node with address %s started with unexpected header", addresses[i])
		}
	}
}

func TestNodeProgress(t *testing.T) {

	// most probably nodes are after block 1 at this point, but to make sure we are waiting
	// for block 1
	// blocks 1-10 are epoch 1 blocks.
	waitForBlock(clients, 1)

	for _, c := range clients {
		epochingClient := epochingtypes.NewQueryClient(c)

		currentEpochResponse, err := epochingClient.CurrentEpoch(
			context.Background(),
			&epochingtypes.QueryCurrentEpochRequest{},
		)

		if err != nil {
			errorString := fmt.Sprintf("Query failed, testnet not running. Error: %v", err)
			panic(errorString)
		}

		if currentEpochResponse.CurrentEpoch != 1 {
			t.Fatalf("Initial epoch should equal 1. Current epoch %d", currentEpochResponse.CurrentEpoch)
		}
	}

	// TODO default epoch interval is equal to 10, we should retrieve it from config
	// block 11 is first block of epoch 2, so if all clients are after block 12, they
	// should be at epoch 2
	waitForBlock(clients, 12)

	for _, c := range clients {
		epochingClient := epochingtypes.NewQueryClient(c)

		currentEpochResponse, err := epochingClient.CurrentEpoch(
			context.Background(),
			&epochingtypes.QueryCurrentEpochRequest{},
		)

		if err != nil {
			errorString := fmt.Sprintf("Query failed, testnet not running. Error: %v", err)
			panic(errorString)
		}

		if currentEpochResponse.CurrentEpoch != 2 {
			t.Errorf("Epoch after 10 blocks, should equal 2. Curent epoch %d", currentEpochResponse.CurrentEpoch)
		}
	}
}

func TestSendTx(t *testing.T) {
	// we are waiting for middle of the epoch to avoid race condidions with bls
	// signer sending transaction and incrementing account sequence numbers
	// which may cause header tx to fail.
	// TODO: Create separate account for sending transactions to avoid race
	// conditions with validator acounts.
	waitForBlock(clients, 14)

	// TODO fix hard coded paths
	node0dataPath := "../.testnets/node0/babylond"
	node0genesisPath := "../.testnets/node0/babylond/config/genesis.json"

	sender, err := NewTestTxSender(node0dataPath, node0genesisPath, clients[0])

	if err != nil {
		panic("failed to init sender")
	}

	tip1, err := sender.getBtcTip()

	if err != nil {
		t.Fatalf("Couldnot retrieve tip")
	}

	res, err := sender.insertNewEmptyHeader(tip1)

	if err != nil {
		t.Fatalf("could not insert new btc header")
	}

	_, err = WaitBtcForHeight(sender.Conn, tip1.Height+1)

	if err != nil {
		t.Log(res.TxResponse)
		t.Fatalf("failed waiting for btc lightclient block")
	}

	tip2, err := sender.getBtcTip()

	if err != nil {
		t.Fatalf("Couldnot retrieve tip")
	}

	if tip2.Height != tip1.Height+1 {
		t.Fatalf("Light client should progress by 1 one block")
	}
}

func getCheckpoint(t *testing.T, conn *grpc.ClientConn, epoch uint64) *checkpointingtypes.RawCheckpointWithMeta {
	queryCheckpoint := checkpointingtypes.NewQueryClient(conn)

	res, err := queryCheckpoint.RawCheckpoint(
		context.Background(),
		checkpointingtypes.NewQueryRawCheckpointRequest(epoch),
	)

	if err != nil {
		t.Fatalf("Failed to retrieve epoch %d", epoch)
	}

	return res.RawCheckpoint
}

func TestSubmitCheckPoint(t *testing.T) {
	node0dataPath := "../.testnets/node0/babylond"
	node0genesisPath := "../.testnets/node0/babylond/config/genesis.json"

	// We are at least on 2 epoch due to `TestNodeProgress` test. At this point
	// checkpoint for epoch 1 should already be sealed
	testEpoch := uint64(1)

	sender, err := NewTestTxSender(node0dataPath, node0genesisPath, clients[0])

	if err != nil {
		panic("failed to init sender")
	}

	rawCheckpoint := getCheckpoint(t, clients[0], testEpoch)

	if rawCheckpoint.Status != checkpointingtypes.Sealed {
		t.Fatalf("Expected checkpoint for epoch %d to be Sealed", testEpoch)
	}

	rawBtcCheckpoint, err := checkpointingtypes.FromRawCkptToBTCCkpt(
		rawCheckpoint.Ckpt,
		sender.getSenderAddress().Bytes(),
	)

	if err != nil {
		t.Fatalf("Could not create raw btc checkpoint from raw chekpoint")
	}

	p1, p2 := txformat.MustEncodeCheckpointData(
		txformat.BabylonTag(txformat.DefautTestTagStr),
		txformat.CurrentVersion,
		rawBtcCheckpoint,
	)

	currentTip, err := sender.getBtcTip()

	if err != nil {
		t.Fatalf("Could not retrieve btc tip")
	}

	firstSubmission := datagen.CreateBlockWithTransaction(currentTip.Header.ToBlockHeader(), p1)

	secondSubmission := datagen.CreateBlockWithTransaction(firstSubmission.HeaderBytes.ToBlockHeader(), p2)

	// first insert all headers
	hresp1, err := sender.insertNewHeader(firstSubmission.HeaderBytes)

	if err != nil {
		t.Fatalf("Could not insert first header")
	}

	_, err = WaitBtcForHeight(clients[0], currentTip.Height+1)

	if err != nil {
		t.Log(hresp1.TxResponse)
		t.Fatalf("failed waiting for btc lightclient block")
	}

	hresp2, err := sender.insertNewHeader(secondSubmission.HeaderBytes)

	if err != nil {
		t.Fatalf("Could not insert second header")
	}

	_, err = WaitBtcForHeight(clients[0], currentTip.Height+2)

	if err != nil {
		t.Log(hresp2.TxResponse)
		t.Fatalf("failed waiting for btc lightclient block")
	}

	// At this point light client chain should be 3 long and inserting spv proofs
	// should succeed
	checkPointInsertResponse, err := sender.insertSpvProof(firstSubmission.SpvProof, secondSubmission.SpvProof)

	if err != nil {
		t.Log(checkPointInsertResponse.TxResponse)
		t.Fatalf("failed to send spv proof")
	}

	err = WaitForNextBlock(clients[0])

	if err != nil {
		t.Fatalf("failed to wait for next babylon block")
	}

	rawCheckpoint = getCheckpoint(t, clients[0], testEpoch)

	if rawCheckpoint.Status != checkpointingtypes.Submitted {
		t.Fatalf("Expected checkpoint for epoch %d to be submitted", testEpoch)
	}
}
