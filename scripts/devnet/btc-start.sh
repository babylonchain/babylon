#!/bin/bash -eux

# USAGE:
# ./btc-start

# Starts an btc chain with a new mining addr.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# These options can be overridden by env
CHAIN_DIR="${CHAIN_DIR:-$CWD/node-data}"
BTC_HOME="${BTC_HOME:-$CHAIN_DIR/btc}"
CLEANUP="${CLEANUP:-1}"

echo "--- Chain Dir = $CHAIN_DIR"
echo "--- BTC HOME = $BTC_HOME"


if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  BTC_HOME=$BTC_HOME $CWD/kill-btc-process.sh

  rm -rf $BTC_HOME
  echo "Removed $BTC_HOME"
fi

btcConf=$BTC_HOME/btcwallet.conf
btcCertPath=$BTC_HOME/certs
btcWalletRpcCert=$btcCertPath/rpc-wallet.cert
btcWalletRpcKey=$btcCertPath/rpc-wallet.key
btcRpcCert=$btcCertPath/rpc.cert
btcRpcKey=$btcCertPath/rpc.key

btcpidPath="$BTC_HOME/pid"
btcdpid="$btcpidPath/btcd.pid"
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

# cp $CWD/conf/sample-btcwallet.conf $btcConf

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

gencerts -d $btcCertPath -H host.docker.internal


echo "1. btcd start"
btcd --simnet --rpclisten 127.0.0.1:18556 --datadir $BTC_HOME/btc-data-dir $flagRpcs $flagRpcBtcCert --rpckey $btcRpcKey --logdir $btcLogs/btcd1 >> $btcLogs/btcd1.log &
# Gets the btc pid
echo $! > $btcdpid
sleep 1
echo "1. btcd finish"

# Creates the wallet
echo "2. btcwallet start"
script -q -c 'btcwallet --simnet -u rpcuser -P rpcpass --rpccert '$btcWalletRpcCert' --rpckey '$btcWalletRpcKey' --cafile '$btcRpcCert' --logdir '$btcLogs'/btcwallet --appdata '$BTC_HOME'/appdata --create' <<ENDDOC /dev/null
walletpass
walletpass
n
n
OK
ENDDOC
sleep 1
echo "2. btcwallet finish"


echo "3. btcwallet start"
btcwallet --simnet --rpclisten=127.0.0.1:18554 $flagRpcUserPass $flagRpcWalletCertKey $flagCatFile --logdir $btcLogs/btcwallet2 >> $btcLogs/btcwallet2.log &
echo $! > $btcwalletpid
sleep 1
echo "3. btcwallet finish"


echo "4. btcctl new mining addr start"
newMiningAddr=$(btcctl --simnet --wallet $flagRpcs $flagRpcWalletCert --rpcserver 127.0.0.1 getnewaddress)
echo "4. btcctl new mining addr finish"
echo "."
echo "."
echo "new mining addr" $newMiningAddr

echo "kills the btcprocess"
pid_value=$(cat "$btcdpid")
kill -s 15 "$pid_value"

echo "starts the btc process again with mining addr" $newMiningAddr
btcd --simnet --rpclisten 127.0.0.1:18556 --miningaddr $newMiningAddr --datadir $BTC_HOME/btc-data-dir $flagRpcs $flagRpcBtcCert --rpckey $btcRpcKey --logdir $btcLogs/btcd2 >> $btcLogs/btcd2.log &
echo $! > $btcdpid
sleep 4

btcctl --simnet --wallet $flagRpcs $flagRpcWalletCert generate 100
echo "..."
echo "generated 100 blocks"
