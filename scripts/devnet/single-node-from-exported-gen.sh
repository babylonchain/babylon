#!/bin/bash -eux

# USAGE:
# ./single-gen.sh <option of full path to babylond>

# Starts an babylon chain with only a single node chain.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

NODE_BIN="${1:-$CWD/../../build/babylond}"

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-test-2}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
DENOM="${DENOM:-ubbn}"
BTC_BASE_HEADER_FILE="${BTC_BASE_HEADER_FILE:-""}"
COVD_HOME="${COVD_HOME:-$CHAIN_DIR/covd}"
covdPKs=$COVD_HOME/pks.json

COVENANT_PK_FILE="${COVENANT_PK_FILE:-"$covdPKs"}"
COVENANT_QUORUM="${COVENANT_QUORUM:-1}"
EXPORTED_GEN_FILE="${EXPORTED_GEN_FILE:-$CHAIN_DIR/test-1/n0/config/genesis.exported.json}"

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
n0cfgDir="$n0dir/config"

# Process id of node 0
n0pid="$hdir/n0.pid"

BTC_BASE_HEADER_FILE=$BTC_BASE_HEADER_FILE COVENANT_PK_FILE=$COVENANT_PK_FILE COVENANT_QUORUM=$COVENANT_QUORUM CHAIN_ID=$CHAIN_ID CHAIN_DIR=$CHAIN_DIR DENOM=$DENOM $CWD/setup-single-node.sh

newGen=$n0cfgDir/genesis.json
tmpGen=$n0cfgDir/tmp_genesis.json
# Replaces values in genesis

inputFile=$n0cfgDir/input.json

# cat $EXPORTED_GEN_FILE | jq .app_state.btclightclient.btc_headers[-1] > $inputFile
# jq '.app_state.btclightclient.btc_headers = [input]' $newGen $inputFile > $tmpGen
# mv $tmpGen $newGen
cat $EXPORTED_GEN_FILE | jq .app_state.btclightclient.btc_headers > $inputFile
jq '.app_state.btclightclient.btc_headers = input' $newGen $inputFile > $tmpGen
mv $tmpGen $newGen

cat $EXPORTED_GEN_FILE | jq .app_state.btcstaking.finality_providers > $inputFile
jq '.app_state.btcstaking.finality_providers = input' $newGen $inputFile > $tmpGen
mv $tmpGen $newGen

cat $EXPORTED_GEN_FILE | jq .app_state.btcstaking.btc_delegations > $inputFile
jq '.app_state.btcstaking.btc_delegations = input' $newGen $inputFile > $tmpGen
mv $tmpGen $newGen

log_path=$hdir/n0.log

$NODE_BIN $home0 start --api.enable true --grpc.address="0.0.0.0:9090" --api.enabled-unsafe-cors --grpc-web.enable=true --log_level info > $log_path 2>&1 &

echo $! > $n0pid

Start the instance
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
