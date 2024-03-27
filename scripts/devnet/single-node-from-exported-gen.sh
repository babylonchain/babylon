#!/bin/bash -eu

# USAGE:
# ./single-node-from-exported-gen.sh <option of full path to babylond>

# Starts an babylon chain getting the data from an exported genesis.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

NODE_BIN="${1:-$CWD/../../build/babylond}"

CHAIN_ID="${CHAIN_ID:-test-2}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
EXPORTED_GEN_FILE="${EXPORTED_GEN_FILE:-$CHAIN_DIR/test-1/n0/config/genesis.exported.json}"

echo "--- Chain ID = $CHAIN_ID"
echo "--- Chain Dir = $CHAIN_DIR"

if [ ! -f $NODE_BIN ]; then
  echo "$NODE_BIN does not exists. build it first with $~ make"
  exit 1
fi

hdir="$CHAIN_DIR/$CHAIN_ID"

# Folder for node
n0dir="$hdir/n0"

# Home flag for folder
home0="--home $n0dir"
n0cfgDir="$n0dir/config"

# Process id of node 0
n0pid="$hdir/n0.pid"

CHAIN_ID=$CHAIN_ID CHAIN_DIR=$CHAIN_DIR $CWD/setup-single-node.sh

newGen=$n0cfgDir/genesis.json
tmpGen=$n0cfgDir/tmp_genesis.json
inputFile=$n0cfgDir/input.json

# TODO: create func
# Replaces values in genesis
cat $EXPORTED_GEN_FILE | jq .app_state.btclightclient.btc_headers > $inputFile
jq '.app_state.btclightclient.btc_headers = input' $newGen $inputFile > $tmpGen
mv $tmpGen $newGen

cat $EXPORTED_GEN_FILE | jq .app_state.btcstaking.finality_providers > $inputFile
jq '.app_state.btcstaking.finality_providers = input' $newGen $inputFile > $tmpGen
mv $tmpGen $newGen

cat $EXPORTED_GEN_FILE | jq .app_state.btcstaking.btc_delegations > $inputFile
jq '.app_state.btcstaking.btc_delegations = input' $newGen $inputFile > $tmpGen
mv $tmpGen $newGen

cat $EXPORTED_GEN_FILE | jq .app_state.btcstaking.params > $inputFile
jq '.app_state.btcstaking.params = input' $newGen $inputFile > $tmpGen
mv $tmpGen $newGen

log_path=$hdir/n0.log
$NODE_BIN $home0 start --api.enable true --grpc.address="0.0.0.0:9090" --api.enabled-unsafe-cors --grpc-web.enable=true --log_level info > $log_path 2>&1 &

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
