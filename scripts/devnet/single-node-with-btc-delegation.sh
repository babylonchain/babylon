#!/bin/bash -eux

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 || exit ; pwd -P )"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
VIGILANTE_HOME="${VIGILANTE_HOME:-$CHAIN_DIR/vigilante}"
CLEANUP="${CLEANUP:-1}"

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

gen_blocks () {
  while true; do
    btcctl --simnet --wallet $flagRpcs $flagRpcWalletCert generate 1
    sleep 10
  done
}

BTC_BASE_HEADER_FILE=$VIGILANTE_HOME/btc-base-header.json
vigilanteConf=$VIGILANTE_HOME/vigilante.yml
fVigConf="--config $vigilanteConf"

# Starts BTC
CHAIN_DIR=$CHAIN_DIR $CWD/btc-start.sh
sleep 2

# Setup Vigilante Conf, just to get a good base BTC header to set in babylon genesis
CLEANUP=0 CHAIN_DIR=$CHAIN_DIR $CWD/vigilante-setup-conf.sh
baseBtcHeader=$(vigilante helpers btc-base-header $fVigConf 0 | jq -r)
echo "$baseBtcHeader" > $BTC_BASE_HEADER_FILE

# Starts the blockchain
BTC_BASE_HEADER_FILE=$BTC_BASE_HEADER_FILE CHAIN_DIR=$CHAIN_DIR $CWD/single-node.sh

CLEANUP=1 CHAIN_DIR=$CHAIN_DIR $CWD/vigilante-start.sh



