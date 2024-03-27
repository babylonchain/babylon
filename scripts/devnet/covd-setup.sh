#!/bin/bash -eu

# USAGE:
# ./covd-setup

# it setups the covenant init files for single node chain

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

CHAIN_ID="${CHAIN_ID:-test-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
COVD_HOME="${COVD_HOME:-$CHAIN_DIR/covd}"
CLEANUP="${CLEANUP:-1}"

if ! command -v jq &> /dev/null
then
  echo "⚠️ jq command could not be found!"
  echo "Install it by checking https://stedolan.github.io/jq/download/"
  exit 1
fi

if ! command -v covd &> /dev/null
then
  echo "⚠️ covd command could not be found!"
  echo "Install it by checking https://github.com/babylonchain/covenant-emulator"
  exit 1
fi

homeF="--home $COVD_HOME"
keyName="covenant"

cfg="$COVD_HOME/covd.conf"
covdPubFile=$COVD_HOME/keyring-test/$keyName.pubkey.json
covdPKs=$COVD_HOME/pks.json

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$COVD_HOME/*.pid $CWD/kill-process.sh

  rm -rf $COVD_HOME
  echo "Removed $COVD_HOME"
fi

covd init $homeF

perl -i -pe 's|ChainID = chain-test|ChainID = "'$CHAIN_ID'"|g' $cfg
perl -i -pe 's|Key = covenant-key|Key = "'$keyName'"|g' $cfg
perl -i -pe 's|Port = 2112|Port = 2115|g' $cfg # any other available port.

covenantPubKey=$(covd create-key --key-name $keyName --chain-id $CHAIN_ID $homeF | jq -r)
echo $covenantPubKey > $covdPubFile

# pub-key, jq does not like -
convenantPk=$(cat $covdPubFile | jq .[] | jq --slurp '.[1]')
echo "[$convenantPk]" > $covdPKs