#!/bin/bash -eux

# USAGE:
# ./fpd-start

# it starts the finality provider for single node chain and validator

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-test-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
FPD_HOME="${FPD_HOME:-$CHAIN_DIR/fpd}"
CLEANUP="${CLEANUP:-1}"

if ! command -v fpd &> /dev/null
then
  echo "⚠️ fpd command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/finality-provider/blob/dev/docs/finality-provider.md"
  exit 1
fi

if ! command -v fpcli &> /dev/null
then
  echo "⚠️ fpcli command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/finality-provider/blob/dev/docs/finality-provider.md"
  exit 1
fi

if ! command -v babylond &> /dev/null
then
  echo "⚠️ babylond command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/babylon"
  exit 1
fi

n0dir="$CHAIN_DIR/$CHAIN_ID/n0"
listenAddr="127.0.0.1:12583"

homeF="--home $FPD_HOME"
cid="--chain-id $CHAIN_ID"
dAddr="--daemon-address $listenAddr"
cfg="$FPD_HOME/fpd.conf"
outdir="$FPD_HOME/out"
logdir="$FPD_HOME/logs"
fpKeyName="keys-finality-provider"

# babylon node Home flag for folder
n0dir="$CHAIN_DIR/$CHAIN_ID/n0"
homeN0="--home $n0dir"
kbt="--keyring-backend test"

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$FPD_HOME/*.pid $CWD/kill-process.sh

  rm -rf $FPD_HOME
  echo "Removed $FPD_HOME"
fi

mkdir -p $outdir
mkdir -p $logdir

# Creates and modifies config
fpd init $homeF --force

perl -i -pe 's|DBPath = '$HOME'/.fpd/data|DBPath = "'$FPD_HOME/data'"|g' $cfg
perl -i -pe 's|ChainID = chain-test|ChainID = "'$CHAIN_ID'"|g' $cfg
perl -i -pe 's|BitcoinNetwork = signet|BitcoinNetwork = simnet|g' $cfg
perl -i -pe 's|Port = 2112|Port = 2734|g' $cfg
perl -i -pe 's|RpcListener = 127.0.0.1:12581|RpcListener = "'$listenAddr'"|g' $cfg

fpd keys add --key-name $fpKeyName $cid $homeF > $outdir/keys-add-keys-finality-provider.txt

pid_file=$FPD_HOME/fpd.pid
fpd start --rpc-listener $listenAddr $homeF > $logdir/fpd-start.log 2>&1 &
echo $! > $pid_file
sleep 2

createFPFile=$outdir/create-finality-provider.json
fpcli create-finality-provider --key-name $fpKeyName $cid $homeF $dAddr --moniker val-fp > $createFPFile
btcPKHex=$(cat $createFPFile | jq '.btc_pk_hex' -r)


fpBbnAddr=$(babylond $homeF keys show $fpKeyName -a $kbt)
# transfer funds to the acc created
babylond tx bank send user $fpBbnAddr 100000000ubbn $homeN0 $kbt $cid -y

registerFPFile=$outdir/register-finality-provider.json
fpcli register-finality-provider $dAddr --btc-pk $btcPKHex > $registerFPFile
