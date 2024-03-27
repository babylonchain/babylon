#!/bin/bash -eux

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 || exit ; pwd -P )"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
VIGILANTE_HOME="${VIGILANTE_HOME:-$CHAIN_DIR/vigilante}"
CLEANUP="${CLEANUP:-1}"
COVD_HOME="${COVD_HOME:-$CHAIN_DIR/covd}"
CHAIN_ID="${CHAIN_ID:-test-1}"

# Cleans everything
if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  $CWD/kill-all-process.sh

  rm -rf $CHAIN_DIR
  echo "Removed $CHAIN_DIR"
fi

if ! command -v vigilante &> /dev/null
then
  echo "⚠️ vigilante command could not be found!"
  echo "Install it by checking https://github.com/Babylonchain/vigilante"
  exit 1
fi

BTC_BASE_HEADER_FILE=$VIGILANTE_HOME/btc-base-header.json
vigilanteConf=$VIGILANTE_HOME/vigilante-submitter.yml
fVigConf="--config $vigilanteConf"

# Starts everything with btc delegation
$CWD/single-node-with-btc-delegation.sh

WAIT_UNTIL=1
amountActiveDels=0
while [ $amountActiveDels -lt $WAIT_UNTIL ]
do
  amountActiveDels="$(babylond q btcstaking btc-delegations active -o json | jq '.btc_delegations | length')"
  echo "Current active dels: $amountActiveDels, waiting to reach $WAIT_UNTIL"
  sleep 10
done

# Kills the running node
babylonChainChain1="$CHAIN_DIR/$CHAIN_ID"
test1n0dir="$babylonChainChain1/n0"
PATH_OF_PIDS=$babylonChainChain1/*.pid $CWD/kill-process.sh

exportedGenFile=$test1n0dir/config/genesis.exported.json

# Export the genesis
babylond --home $test1n0dir export > $exportedGenFile

# Updates the chain id
CHAIN_ID=test-2
# Starts everything from a new chain id
CHAIN_ID=$CHAIN_ID EXPORTED_GEN_FILE=$exportedGenFile $CWD/single-node-from-exported-gen.sh

WAIT_UNTIL=1
amountActiveDels=0
while [ $amountActiveDels -lt $WAIT_UNTIL ]
do
  amountActiveDels="$(babylond q btcstaking btc-delegations active -o json | jq '.btc_delegations | length')"
  echo "Current active dels: $amountActiveDels, waiting to reach $WAIT_UNTIL"
  sleep 10
done
echo "FINALLY STARTED CHAIN 2 WITH BTC DELS"