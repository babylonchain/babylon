#!/bin/bash -eu

# USAGE:
# ./vigilante-setup-conf.sh

# Creates vigilante config file.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

CHAIN_ID="${CHAIN_ID:-test-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
CHAIN_HOME="$CHAIN_DIR/$CHAIN_ID"

N0_HOME="${N0_HOME:-$CHAIN_HOME/n0}"
BTC_HOME="${BTC_HOME:-$CHAIN_DIR/btc}"
VIGILANTE_HOME="${VIGILANTE_HOME:-$CHAIN_DIR/vigilante}"
LISTEN_PORT="${LISTEN_PORT:-8067}"
SERVER_PORT="${SERVER_PORT:-2135}"

BTC_RPC_CERT="${BTC_RPC_CERT:-$BTC_HOME/certs/rpc.cert}"
BTC_WALLET_RPC_CERT="${BTC_WALLET_RPC_CERT:-$BTC_HOME/certs/rpc-wallet.cert}"
SUBMITTER_ADDR="${SUBMITTER_ADDR:-bbn1dnug7399p0xg4x2ccduegu94gxshrrl78r8mz6}"
CONF_PATH="${CONF_PATH:-$VIGILANTE_HOME/vigilante-submitter.yml}"

CLEANUP="${CLEANUP:-1}"

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$VIGILANTE_HOME/pid/*.pid $CWD/kill-process.sh

  rm -rf $VIGILANTE_HOME
  echo "Removed $VIGILANTE_HOME"
fi

mkdir -p $VIGILANTE_HOME

echo "
common:
  log-format: "auto" # format of the log (json|auto|console|logfmt)
  log-level: "debug" # log level (debug|warn|error|panic|fatal)
  retry-sleep-time: 5s
  max-retry-sleep-time: 5m
btc:
  no-client-tls: false # use true for bitcoind as it does not support tls
  ca-file: $BTC_RPC_CERT
  endpoint: 127.0.0.1:18556 # use port 18443 for bitcoind regtest
  estimate-mode: CONSERVATIVE # only needed by bitcoind
  tx-fee-max: 20000 # maximum tx fee, 20,000sat/kvb
  tx-fee-min: 1000 # minimum tx fee, 1,000sat/kvb
  default-fee: 10000 # 1,000sat/kvb
  target-block-num: 2
  wallet-endpoint: 127.0.0.1:18554
  wallet-password: walletpass
  wallet-name: default
  wallet-lock-time: 10
  wallet-ca-file: $BTC_WALLET_RPC_CERT
  net-params: simnet  # use regtest for bitcoind as it does not support simnet
  username: rpcuser
  password: rpcpass
  reconnect-attempts: 3
  btc-backend: btcd # {btcd, bitcoind}
  zmq-endpoint: ~  # use tcp://127.0.0.1:29000 if btc-backend is bitcoind
babylon:
  key: submitter
  chain-id: $CHAIN_ID
  rpc-addr: http://localhost:26657
  grpc-addr: https://localhost:9090
  account-prefix: bbn
  keyring-backend: test
  gas-adjustment: 1.2
  gas-prices: 20ubbn
  key-directory: $N0_HOME
  debug: true
  timeout: 20s
  block-timeout: ~
  output-format: json
  submitter-address: $SUBMITTER_ADDR
  sign-mode: direct
grpc:
  onetime-tls-key: true
  rpc-key: \"\"
  rpc-cert: $VIGILANTE_HOME/rpc.cert
  endpoints:
    - localhost:$LISTEN_PORT
grpcweb:
  placeholder: grpcwebconfig
metrics:
  host: 0.0.0.0
  server-port: $SERVER_PORT
submitter:
  netparams: simnet
  buffer-size: 10
  resubmit-fee-multiplier: 1
  polling-interval-seconds: 60
  resend-interval-seconds: 1800
reporter:
  netparams: simnet
  btc_cache_size: 1000
  max_headers_in_msg: 100
monitor:
  checkpoint-buffer-size: 1000
  btc-block-buffer-size: 1000
  btc-cache-size: 1000
  btc-confirmation-depth: 6
  liveness-check-interval-seconds: 100
  max-live-btc-heights: 200
  enable-liveness-checker: true
  enable-slasher: true
  btcnetparams: simnet
btcstaking-tracker:
  check-delegations-interval: 1m
  delegations-batch-size: 100
  check-if-delegation-active-interval: 5m
  retry-submit-unbonding-interval: 1m
  max-jitter-interval: 30s
  btcnetparams: simnet
" > $CONF_PATH
