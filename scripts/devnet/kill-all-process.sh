#!/bin/bash

# USAGE:
# ./kill-all-process.sh

# Kill all the process stored in the PID paths of possible generated processes in CHAIN_DIR

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 || exit ; pwd -P )"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
CHAIN_ID="${CHAIN_ID:-test-1}"

babylonChain="$CHAIN_DIR/$CHAIN_ID"
babylonChain2="$CHAIN_DIR/test-2"
BTC_HOME="${BTC_HOME:-$CHAIN_DIR/btc}"
VIGILANTE_HOME="${VIGILANTE_HOME:-$CHAIN_DIR/vigilante}"
COVD_HOME="${COVD_HOME:-$CHAIN_DIR/covd}"
EOTS_HOME="${EOTS_HOME:-$CHAIN_DIR/eots}"
FPD_HOME="${FPD_HOME:-$CHAIN_DIR/fpd}"
BTC_STAKER_HOME="${BTC_STAKER_HOME:-$CHAIN_DIR/btc-staker}"

PATH_OF_PIDS=$babylonChain/*.pid $CWD/kill-process.sh
PATH_OF_PIDS=$babylonChain2/*.pid $CWD/kill-process.sh
PATH_OF_PIDS=$VIGILANTE_HOME/pid/*.pid $CWD/kill-process.sh
PATH_OF_PIDS=$BTC_HOME/pid/*.pid $CWD/kill-process.sh
PATH_OF_PIDS=$COVD_HOME/*.pid $CWD/kill-process.sh
PATH_OF_PIDS=$EOTS_HOME/*.pid $CWD/kill-process.sh
PATH_OF_PIDS=$FPD_HOME/*.pid $CWD/kill-process.sh
PATH_OF_PIDS=$BTC_STAKER_HOME/pid/*.pid $CWD/kill-process.sh