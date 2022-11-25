#!/bin/bash
# microtick and bitcanna contributed significantly here.
# Pebbledb state sync script.
# invoke like: bash statesync.bash

## USAGE RUNDOWN
# Not for use on live nodes
# For use when testing.
# Assumes that ~/.evmosd doesn't exist
# can be modified to suit your purposes if ~/.evmosd does already exist

set -uxe

# Set Golang environment variables.
export GOPATH=~/go
export PATH=$PATH:~/go/bin

# Install with pebbledb
go mod edit -replace github.com/tendermint/tm-db=github.com/baabeetaa/tm-db@pebble
go mod tidy
make build

# Install with goleveldb
# go install ./...

# NOTE: ABOVE YOU CAN USE ALTERNATIVE DATABASES, HERE ARE THE EXACT COMMANDS
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb' -tags rocksdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb' -tags badgerdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb' -tags boltdb ./...

# Initialize chain.
babylond init test

# Get Genesis
curl http://node.mainnet.babylonchain.io:26657/genesis | jq .result.genesis >~/.babylond/config/genesis.json

# Get "trust_hash" and "trust_height".
INTERVAL=100
LATEST_HEIGHT=$(curl -s http://node.mainnet.babylonchain.io:26657/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$(($LATEST_HEIGHT - $INTERVAL))
TRUST_HASH=$(curl -s "http://node.mainnet.babylonchain.io:26657/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

# Print out block and transaction hash from which to sync state.
echo "trust_height: $BLOCK_HEIGHT"
echo "trust_hash: $TRUST_HASH"

# Export state sync variables.
export BABYLOND_STATESYNC_ENABLE=true
export BABYLOND_P2P_MAX_NUM_OUTBOUND_PEERS=200
export BABYLOND_STATESYNC_RPC_SERVERS="http://node.mainnet.babylonchain.io:26657,http://node.mainnet.babylonchain.io:26657"
export BABYLOND_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export BABYLOND_STATESYNC_TRUST_HASH=$TRUST_HASH
# Export state sync variables.
export BABYLOND_P2P_LADDR=tcp://0.0.0.0:2100
export BABYLOND_RPC_LADDR=tcp://0.0.0.0:2101
export BABYLOND_GRPC_ADDRESS=127.0.0.1:2102
export BABYLOND_API_ADDRESS=tcp://127.0.0.1:2103
export BABYLOND_NODE=tcp://127.0.0.1:2101





# Fetch and set list of seeds from chain registry.
export BABYLOND_P2P_PERSISTENT_PEERS=226ad8bc898ff89449032b73458768f646aa30a3@node.mainnet.babylonchain.io:26656

# Start chain.
babylond start --x-crisis-skip-assert-invariants
