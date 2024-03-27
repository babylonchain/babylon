#!/bin/bash -eu

# USAGE:
# ./btc-start

# Starts an btc chain with a new mining addr.
# Btc processes needs sleep timing --"

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# These options can be overridden by env
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
BTC_HOME="${BTC_HOME:-$CHAIN_DIR/btc}"
CLEANUP="${CLEANUP:-1}"

echo "--- Chain Dir = $CHAIN_DIR"
echo "--- BTC HOME = $BTC_HOME"

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$BTC_HOME/pid/*.pid $CWD/kill-process.sh
  sleep 3 # takes some time to kill the process and start again...

  rm -rf $BTC_HOME
  echo "Removed $BTC_HOME"
fi

btcCertPath=$BTC_HOME/certs
btcWalletRpcCert=$btcCertPath/rpc-wallet.cert
btcWalletRpcKey=$btcCertPath/rpc-wallet.key
btcRpcCert=$btcCertPath/rpc.cert
btcRpcKey=$btcCertPath/rpc.key

btcpidPath="$BTC_HOME/pid"
btcdpid="$btcpidPath/btcd.pid"
genblockspid="$btcpidPath/genblocks.pid"
btcwalletpid="$btcpidPath/btcwallet.pid"
btcLogs="$BTC_HOME/logs"

mkdir -p $BTC_HOME
mkdir -p $btcLogs
mkdir -p $btcpidPath
mkdir -p $btcCertPath

flagRpcs="--rpcuser=rpcuser --rpcpass=rpcpass"
flagCatFile="--cafile=$btcRpcCert"
flagRpcBtcCert="--rpccert $btcRpcCert"
flagRpcWalletCert="--rpccert $btcWalletRpcCert"
flagRpcWalletCertKey="$flagRpcWalletCert --rpckey $btcWalletRpcKey"
flagRpcUserPass="-u rpcuser -P rpcpass"

if ! command -v btcd &> /dev/null
then
  echo "⚠️ btcd command could not be found!"
  echo "Install it by checking https://github.com/btcsuite/btcd"
  exit 1
fi

if ! command -v btcwallet &> /dev/null
then
  echo "⚠️ btcwallet command could not be found!"
  echo "Install it by checking https://github.com/btcsuite/btcwallet"
  exit 1
fi

if ! command -v btcctl &> /dev/null
then
  echo "⚠️ btcctl command could not be found!"
  echo "Install it by checking https://github.com/btcsuite/btcd"
  exit 1
fi

if ! command -v gencerts &> /dev/null
then
  echo "⚠️ gencerts command could not be found!"
  echo "Install it by checking https://github.com/gerrywastaken/gencert/releases/tag/v0.1.6"
  exit 1
fi

if ! command -v jq &> /dev/null
then
  echo "⚠️ jq command could not be found!"
  echo "Install it by checking https://stedolan.github.io/jq/download/"
  exit 1
fi

gen_blocks () {
  echo "1 block generated each 8s"

  while true; do
    btcctl --simnet --wallet $flagRpcs $flagRpcWalletCert generate 1 > /dev/null 2>&1
    sleep 8
  done
}

gencerts -d $btcCertPath -H host.docker.internal


btcd --simnet --rpclisten 127.0.0.1:18556 --datadir $BTC_HOME/btc-data-dir $flagRpcs $flagRpcBtcCert --rpckey $btcRpcKey --logdir $btcLogs/btcd-to-create-wallet >> $btcLogs/btcd-to-create-wallet.log &
echo $! > $btcdpid
sleep 1

# Creates the wallet
script -q -c 'btcwallet --simnet -u rpcuser -P rpcpass --rpccert '$btcWalletRpcCert' --rpckey '$btcWalletRpcKey' --cafile '$btcRpcCert' --logdir '$btcLogs'/btc-create-wallet --appdata '$BTC_HOME'/appdata-wallet --create' <<ENDDOC /dev/null
walletpass
walletpass
n
n
OK
ENDDOC
sleep 1

btcwallet --simnet --rpclisten=127.0.0.1:18554 --rpcconnect=127.0.0.1:18556 --appdata $BTC_HOME/appdata-wallet $flagRpcUserPass $flagRpcWalletCertKey $flagCatFile --logdir $btcLogs/btcwallet2 >> $btcLogs/btcwallet2.log &
echo $! > $btcwalletpid
sleep 1

newMiningAddr=$(btcctl --simnet --wallet $flagRpcs $flagRpcWalletCert --rpcserver 127.0.0.1 getnewaddress)
echo "new mining addr" $newMiningAddr

echo "kills the btcprocess"
pid_value=$(cat "$btcdpid")
kill -s 15 "$pid_value"
sleep 2

echo "starts the btc process again with mining addr" $newMiningAddr
btcd --simnet --rpclisten 127.0.0.1:18556 --miningaddr $newMiningAddr --datadir $BTC_HOME/btc-data-dir $flagRpcs $flagRpcBtcCert --rpckey $btcRpcKey --logdir $btcLogs/btcd2 >> $btcLogs/btcd2.log &
echo $! > $btcdpid
sleep 4

blockHeight=120

btcctl --simnet --wallet $flagRpcs $flagRpcWalletCert setgenerate 0
btcctl --simnet --wallet $flagRpcs $flagRpcWalletCert generate $blockHeight
echo "generated $blockHeight blocks"

# keeps mining 1 block each 8 sec.
gen_blocks &
echo $! > $genblockspid
