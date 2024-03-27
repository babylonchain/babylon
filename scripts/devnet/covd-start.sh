#!/bin/bash -eu

# USAGE:
# ./covd-start

# it starts the covenant for single node chain

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

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

if ! command -v babylond &> /dev/null
then
  echo "⚠️ babylond command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/babylon"
  exit 1
fi

homeF="--home $COVD_HOME"
n0dir="$CHAIN_DIR/$CHAIN_ID/n0"
homeN0="--home $n0dir"
kbt="--keyring-backend test"
cid="--chain-id $CHAIN_ID"

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$COVD_HOME/*.pid $CWD/kill-process.sh

  rm -rf $COVD_HOME
  echo "Removed $COVD_HOME"
fi

if [[ "$SETUP" == 1 || "$SETUP" == "1" ]]; then
  $CWD/covd-setup.sh
fi

# transfer funds to the covenant acc created
covenantAddr=$(babylond $homeF keys show covenant -a $kbt)
babylond tx bank send user $covenantAddr 100000000ubbn $homeN0 $kbt $cid -y

# Start Covenant
covd start $homeF >> $COVD_HOME/covd-start.log &
echo $! > $COVD_HOME/covd.pid
