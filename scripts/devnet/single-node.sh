#!/bin/bash -eux

# USAGE:
# ./single-gen.sh <option of full path to babylond>

# Starts an babylon chain with only a single node chain.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

NODE_BIN="${1:-$CWD/../../build/babylond}"

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-test-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/node-data}"
DENOM="${DENOM:-ubbn}"

echo "--- Chain ID = $CHAIN_ID"
echo "--- Chain Dir = $CHAIN_DIR"
echo "--- Coin Denom = $DENOM"

if [ ! -f $NODE_BIN ]; then
  echo "$NODE_BIN does not exists. build it first with $~ make"
  exit 1
fi

hdir="$CHAIN_DIR/$CHAIN_ID"

# Folder for node
n0dir="$hdir/n0"

# Home flag for folder
home0="--home $n0dir"

# Process id of node 0
n0pid="$n0dir/pid"

CHAIN_ID=$CHAIN_ID CHAIN_DIR=$CHAIN_DIR DENOM=$DENOM $CWD/setup-single-node.sh

log_path=$hdir.n0.log

$NODE_BIN $home0 start --api.enable true --grpc.address="0.0.0.0:9090" --api.enabled-unsafe-cors --grpc-web.enable=true --log_level debug > $log_path 2>&1 &

# Gets the node pid
echo $! > $n0pid

# Start the instance
echo "--- Starting node..."
echo
echo "Logs:"
echo "  * tail -f $log_path"
echo
echo "Env for easy access:"
echo "export H1='--home $n0dir'"
echo
echo "Command Line Access:"
echo "  * $NODE_BIN --home $n0dir status"
