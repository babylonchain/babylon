#!/bin/bash -eux

# USAGE:
# ./vigilante-start

# Starts an btc chain with a new mining addr.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-test-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
CHAIN_HOME="$CHAIN_DIR/$CHAIN_ID"
NODE_BIN="${NODE_BIN:-$CWD/../../build/babylond}"
N0_HOME="${N0_HOME:-$CHAIN_HOME/n0}"
BTC_HOME="${BTC_HOME:-$CHAIN_DIR/btc}"
VIGILANTE_HOME="${VIGILANTE_HOME:-$CHAIN_DIR/vigilante}"
CLEANUP="${CLEANUP:-1}"

echo "--- Chain Dir = $CHAIN_DIR"
echo "--- BTC HOME = $BTC_HOME"

vigilantepidPath="$VIGILANTE_HOME/pid"
vigilanteLogs="$VIGILANTE_HOME/logs"

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$vigilantepidPath/*.pid $CWD/kill-process.sh

  rm -rf $VIGILANTE_HOME
  echo "Removed $VIGILANTE_HOME"
fi

mkdir -p $VIGILANTE_HOME
mkdir -p $vigilantepidPath
mkdir -p $vigilanteLogs

btcCertPath=$BTC_HOME/certs
btcRpcCert=$btcCertPath/rpc.cert
btcWalletRpcCert=$btcCertPath/rpc-wallet.cert

vigilanteConfSub=$VIGILANTE_HOME/vigilante-submitter.yml
vigilanteConfRep=$VIGILANTE_HOME/vigilante-reporter.yml

reporterpid="$vigilantepidPath/reporter.pid"
submitterpid="$vigilantepidPath/submitter.pid"

kbt="--keyring-backend test"
submitterAddr=$($NODE_BIN --home $N0_HOME keys show submitter -a $kbt)

CONF_PATH=$vigilanteConfSub CLEANUP=0 SUBMITTER_ADDR=$submitterAddr $CWD/vigilante-setup-conf.sh
CONF_PATH=$vigilanteConfRep CLEANUP=0 SUBMITTER_ADDR=$submitterAddr SERVER_PORT=2134 LISTEN_PORT=8068 $CWD/vigilante-setup-conf.sh

vigilante --config $vigilanteConfRep reporter > $vigilanteLogs/reporter.log 2>&1 &
echo $! > $reporterpid

vigilante --config $vigilanteConfSub submitter > $vigilanteLogs/submitter.log 2>&1 &
echo $! > $submitterpid