#!/bin/bash

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 || exit ; pwd -P )"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
CHAIN_ID="${CHAIN_ID:-test-1}"

babylonChain="$CHAIN_DIR/$CHAIN_ID"
BTC_HOME="${BTC_HOME:-$CHAIN_DIR/btc}"
VIGILANTE_HOME="${VIGILANTE_HOME:-$CHAIN_DIR/vigilante}"

PATH_OF_PIDS=$babylonChain/*.pid $CWD/kill-process.sh
PATH_OF_PIDS=$VIGILANTE_HOME/pid/*.pid $CWD/kill-process.sh
PATH_OF_PIDS=$BTC_HOME/pid/*.pid $CWD/kill-process.sh