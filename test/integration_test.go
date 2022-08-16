//go:build integration
// +build integration

package babylon_integration_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	appparams "github.com/babylonchain/babylon/app/params"
	bbl "github.com/babylonchain/babylon/types"
	lightclient "github.com/babylonchain/babylon/x/btclightclient/types"
	ec "github.com/babylonchain/babylon/x/epoching/types"
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
			panic("Integration tests failed, due to node failure")
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

	// This is needed so that all address prefixes are in bbl format
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
	// TODO currently btclightclient hardcodes this header in some function. Ultimately
	// we would like to get it from config file, and assert here that each node
	// start with genesis header from this config file
	hardcodedHeaderHash, _ := bbl.NewBTCHeaderHashBytesFromHex("00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47")
	hardcodedHeaderHeight := uint64(736056)

	for i, c := range clients {
		lc := lightclient.NewQueryClient(c)

		res, err := lc.Tip(context.Background(), lightclient.NewQueryTipRequest())

		if err != nil {
			// this is fatal, as it means we most probably did not get any response and
			// at least one of the nodes is down
			t.Fatalf("Test failed due to client error: %v to node with address %s", err, addresses[i])
		}

		if res.Header.Height != hardcodedHeaderHeight || !res.Header.Hash.Eq(&hardcodedHeaderHash) {
			t.Errorf("Node with address %s started with unexpected header", addresses[i])
		}
	}
}

func TestNodeProgres(t *testing.T) {

	// most probably nodes are after block 1 at this point, but to make sure we are waiting
	// for block 1
	// blocks 1-10 are epoch 1 blocks.
	waitForBlock(clients, 1)

	for _, c := range clients {
		epochingClient := ec.NewQueryClient(c)

		currentEpochResponse, err := epochingClient.CurrentEpoch(context.Background(), &ec.QueryCurrentEpochRequest{})

		if err != nil {
			panic("Querry failed, testnet not running")
		}

		if currentEpochResponse.CurrentEpoch != 1 {
			t.Fatalf("Initial epoch should equal 1. Current epoch %d", currentEpochResponse.CurrentEpoch)
		}
	}

	// TODO default epoch interval is equal to 10, we should retrieve it from config
	// block 11 is first block of epoch 2, so if all clients are after block 11, they
	// should be at epoch 2
	waitForBlock(clients, 11)

	for _, c := range clients {
		epochingClient := ec.NewQueryClient(c)

		currentEpochResponse, err := epochingClient.CurrentEpoch(context.Background(), &ec.QueryCurrentEpochRequest{})

		if err != nil {
			panic("Querry failed, testnet not running")
		}

		if currentEpochResponse.CurrentEpoch != 2 {
			t.Errorf("Epoch after 10 blocks, should equal 2. Curent epoch %d", currentEpochResponse.CurrentEpoch)
		}
	}
}
