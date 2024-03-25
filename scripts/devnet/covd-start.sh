#!/bin/bash -eux

# USAGE:
# ./covd-start

# it starts the covenant for single node chain

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

NODE_BIN="${1:-$CWD/../../build/babylond}"

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-test-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
COVD_HOME="${COVD_HOME:-$CHAIN_DIR/covd}"
CLEANUP="${CLEANUP:-1}"
SETUP="${SETUP:-1}"

if ! command -v covd &> /dev/null
then
  echo "⚠️ covd command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/covenant-emulator"
  exit 1
fi

# Home flag for folder
homeF="--home $COVD_HOME"

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$COVD_HOME/*.pid $CWD/kill-process.sh

  rm -rf $COVD_HOME
  echo "Removed $COVD_HOME"
fi

if [[ "$SETUP" == 1 || "$SETUP" == "1" ]]; then
  $CWD/covd-setup.sh
fi


# Start Covenant
covd start $homeF >> $COVD_HOME/covd-start.log &
echo $! > $COVD_HOME/covd.pid
