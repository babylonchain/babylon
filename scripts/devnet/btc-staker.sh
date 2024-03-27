#!/bin/bash -eu

# USAGE:
# ./btc-staker

# Starts an btc staker and sends stake tx to btc.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
CHAIN_ID="${CHAIN_ID:-test-1}"
BTC_HOME="${BTC_HOME:-$CHAIN_DIR/btc}"
BTC_STAKER_HOME="${BTC_STAKER_HOME:-$CHAIN_DIR/btc-staker}"
CLEANUP="${CLEANUP:-1}"

pidPath=$BTC_STAKER_HOME/pid
btcctldOutputDirPath=$BTC_STAKER_HOME/btcctl/output

stakercliDirHome=$BTC_STAKER_HOME/stakecli

stakercliConfigFile=$stakercliDirHome/config.conf
stakercliDataDir=$stakercliDirHome/data
stakercliLogsDir=$stakercliDirHome/logs
stakercliOutputDir=$stakercliDirHome/output
stakercliDBDir=$stakercliDirHome/db

BTC_STAKER_KEY="btc-staker"

btcCertPath=$BTC_HOME/certs
btcRpcCert=$btcCertPath/rpc.cert
btcWalletRpcCert=$btcCertPath/rpc-wallet.cert

n0dir="$CHAIN_DIR/$CHAIN_ID/n0"

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$pidPath/*.pid $CWD/kill-process.sh

  rm -rf $BTC_STAKER_HOME
  echo "Removed $BTC_STAKER_HOME"
fi

mkdir -p $pidPath
mkdir -p $btcctldOutputDirPath
mkdir -p $stakercliLogsDir
mkdir -p $stakercliOutputDir

flagRpcs="--rpcuser=rpcuser --rpcpass=rpcpass"
flagRpcWalletCert="--rpccert $btcWalletRpcCert"

if ! command -v stakercli &> /dev/null
then
  echo "⚠️ stakercli command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/btc-staker"
  exit 1
fi

if ! command -v stakerd &> /dev/null
then
  echo "⚠️ stakerd command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/btc-staker"
  exit 1
fi

if ! command -v jq &> /dev/null
then
  echo "⚠️ jq command could not be found!"
  echo "Install it by checking https://stedolan.github.io/jq/download/"
  exit 1
fi

stakercli admin dump-config --config-file-dir $stakercliConfigFile

#[Application Options]
perl -i -pe 's|StakerdDir = '$HOME'/.stakerd|StakerdDir = "'$stakercliDirHome'"|g' $stakercliConfigFile
perl -i -pe 's|ConfigFile = '$HOME'/.stakerd/stakerd.conf|ConfigFile = "'$stakercliConfigFile'"|g' $stakercliConfigFile
perl -i -pe 's|DataDir = '$HOME'/.stakerd/data|DataDir = "'$stakercliDataDir'"|g' $stakercliConfigFile
perl -i -pe 's|LogDir = '$HOME'/.stakerd/logs|LogDir = "'$stakercliLogsDir'"|g' $stakercliConfigFile
#[walletconfig]
perl -i -pe 's|WalletName = wallet|WalletName = default|g' $stakercliConfigFile
#[walletrpcconfig]
perl -i -pe 's|Host = localhost:18556|Host = 127.0.0.1:18554|g' $stakercliConfigFile
perl -i -pe 's|DisableTls = true|DisableTls = false|g' $stakercliConfigFile
perl -i -pe 's|RPCWalletCert =|RPCWalletCert = "'$btcWalletRpcCert'"|g' $stakercliConfigFile
perl -i -pe 's|RawRPCWalletCert = "'$btcWalletRpcCert'"|RawRPCWalletCert =|g' $stakercliConfigFile
#[chain]
perl -i -pe 's|Network = testnet|Network = simnet|g' $stakercliConfigFile
#[btcd]
perl -i -pe 's|RPCHost = 127.0.0.1:18334|RPCHost = 127.0.0.1:18556|g' $stakercliConfigFile
perl -i -pe 's|RPCUser = user|RPCUser = rpcuser|g' $stakercliConfigFile
perl -i -pe 's|RPCPass = pass|RPCPass = rpcpass|g' $stakercliConfigFile
perl -i -pe 's|RPCCert = '$HOME'/.btcd/rpc.cert|RPCCert = "'$btcRpcCert'"|g' $stakercliConfigFile
#[babylon]
perl -i -pe 's|Key = node0|Key = "'$BTC_STAKER_KEY'"|g' $stakercliConfigFile
perl -i -pe 's|ChainID = chain-test|ChainID = "'$CHAIN_ID'"|g' $stakercliConfigFile
perl -i -pe 's|KeyDirectory = '$HOME'/.stakerd|KeyDirectory = "'$n0dir'"|g' $stakercliConfigFile
#[dbconfig]
perl -i -pe 's|DBPath = '$HOME'/.stakerd/data|DBPath = "'$stakercliDBDir'"|g' $stakercliConfigFile
#[stakerconfig]
# perl -i -pe 's|BabylonStallingInterval = 1m0s|BabylonStallingInterval = 40s|g' $stakercliConfigFile

# starts the staker daemon
stakerd --configfile=$stakercliConfigFile > $stakercliLogsDir/daemon.log 2>&1 &
echo $! > $pidPath/stakerd.pid
sleep 2 # waits for the daemon to load.

finalityProviderBTCPubKey=$(stakercli daemon babylon-finality-providers | jq .finality_providers[0].bitcoin_public_Key -r)
echo $finalityProviderBTCPubKey > $stakercliOutputDir/fpbtc.pub.key

stakerBTCAddrListOutput=$(stakercli daemon list-outputs | jq .outputs[-1].address -r)
echo $stakerBTCAddrListOutput > $stakercliOutputDir/list.output.last.addr

# Creates the btc delegation
stakercli daemon stake --staker-address $stakerBTCAddrListOutput --staking-amount 1000000 --finality-providers-pks $finalityProviderBTCPubKey --staking-time 10000 > $stakercliOutputDir/btc-staking-tx.json

# Generate a few blocks to confirm the tx.
btcctl --simnet --wallet $flagRpcs $flagRpcWalletCert generate 11