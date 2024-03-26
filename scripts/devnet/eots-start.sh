#!/bin/bash -eux

# USAGE:
# ./eots-start

# it starts the covenant for single node chain

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

NODE_BIN="${1:-$CWD/../../build/babylond}"

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-test-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
EOTS_HOME="${EOTS_HOME:-$CHAIN_DIR/eots}"
CLEANUP="${CLEANUP:-1}"

if ! command -v eotsd &> /dev/null
then
  echo "⚠️ eotsd command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/finality-provider/blob/dev/docs/eots.md"
  exit 1
fi

# Home flag for folder
homeF="--home $EOTS_HOME"
cfg="$EOTS_HOME/eotsd.conf"

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$EOTS_HOME/*.pid $CWD/kill-process.sh

  rm -rf $EOTS_HOME
  echo "Removed $EOTS_HOME"
fi

eotsd init $homeF
perl -i -pe 's|DBPath = '$HOME'/.eotsd/data|DBPath = "'$EOTS_HOME/data'"|g' $cfg

# Start Covenant
eotsd start $homeF >> $EOTS_HOME/eots-start.log &
echo $! > $EOTS_HOME/eots.pid
