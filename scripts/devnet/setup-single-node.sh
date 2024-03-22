#!/bin/bash -eux

# USAGE:
# ./setup-single-node.sh <option of full path to babylond>

# it setups the single-node home files base struct

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

NODE_BIN="${1:-$CWD/../../build/babylond}"

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-test-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
DENOM="${DENOM:-ubbn}"
CLEANUP="${CLEANUP:-1}"
LOG_LEVEL="${LOG_LEVEL:-info}"
VOTING_PERIOD="${VOTING_PERIOD:-20s}"
EXPEDITED_VOTING_PERIOD="${EXPEDITED_VOTING_PERIOD:-10s}"
BTC_BASE_HEADER_FILE="${BTC_BASE_HEADER_FILE:-""}"

# Default 1 account keys + 1 user key with no special grants
VAL0_KEY="val"
VAL0_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"
# bbnvaloper1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87yrjr7tn

USER_KEY="user"
USER_MNEMONIC="pony glide frown crisp unfold lawn cup loan trial govern usual matrix theory wash fresh address pioneer between meadow visa buffalo keep gallery swear"

SUBMITTER_KEY="submitter"
SUBMITTER_MNEMONIC="catalog disagree royal alley edge negative erase clip dolphin undo pipe fire small siren bird crowd reopen wrestle stumble survey rib gospel master toilet"

NEWLINE=$'\n'

hdir="$CHAIN_DIR/$CHAIN_ID"

if ! command -v jq &> /dev/null
then
  echo "⚠️ jq command could not be found!"
  echo "Install it by checking https://stedolan.github.io/jq/download/"
  exit 1
fi

echo "--- Chain ID = $CHAIN_ID"
echo "--- Chain Dir = $CHAIN_DIR"
echo "--- Coin Denom = $DENOM"

if [ ! -f $NODE_BIN ]; then
  echo "$NODE_BIN does not exists. Build it first with $~ make"
  exit 1
fi

# Folder for node
n0dir="$hdir/n0"

# Home flag for folder
home0="--home $n0dir"

# Config directories for node
n0cfgDir="$n0dir/config"

# Config files for nodes
n0cfg="$n0cfgDir/config.toml"

# App config file for node
n0app="$n0cfgDir/app.toml"
n0PrivKey="$n0cfgDir/priv_validator_key.json"

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  PATH_OF_PIDS=$hdir/*.pid $CWD/kill-process.sh

  rm -rf $hdir
  echo "Removed $n0dir"
fi

# Common flags
kbt="--keyring-backend test"
cid="--chain-id $CHAIN_ID"

# Check if the data dir has been initialized already
if [[ -d "$hdir" ]]; then
  echo "===================================="
  echo "CONTINUING CHAIN FROM PREVIOUS STATE"
  echo "===================================="

  exit 0
fi

echo "====================================="
echo "STARTING NEW CHAIN WITH GENESIS STATE"
echo "====================================="

echo "--- Creating $NODE_BIN validator with chain-id=$CHAIN_ID"

# Build genesis file and create accounts
coins="1000000000000$DENOM"
coins_user="1000000000000$DENOM"

echo "--- Initializing home..."

# Initialize the home directory of node
$NODE_BIN $home0 $cid init n0 &>/dev/null

echo "--- Importing keys..."
echo "$VAL0_MNEMONIC$NEWLINE"
yes "$VAL0_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $VAL0_KEY $kbt --recover
yes "$USER_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $USER_KEY $kbt --recover
yes "$SUBMITTER_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $SUBMITTER_KEY $kbt --recover

echo "--- Adding addresses..."
$NODE_BIN $home0 keys show $VAL0_KEY -a $kbt
$NODE_BIN $home0 keys show $VAL0_KEY -a --bech val $kbt
$NODE_BIN $home0 keys show $USER_KEY -a $kbt # bbnvaloper1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87yrjr7tn

VAL0_ADDR=$($NODE_BIN $home0 keys show $VAL0_KEY -a $kbt --bech val)

$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $VAL0_KEY -a $kbt) $coins &>/dev/null
$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $USER_KEY -a $kbt) $coins_user &>/dev/null
$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $SUBMITTER_KEY -a $kbt) $coins_user &>/dev/null
$NODE_BIN $home0 create-bls-key $($NODE_BIN $home0 keys show $VAL0_KEY -a $kbt)

  # | .initial_height="1"
echo "--- Patching genesis..."
jq '.consensus_params["block"]["time_iota_ms"]="5000"
  | .app_state["crisis"]["constant_fee"]["denom"]="'$DENOM'"
  | .app_state["mint"]["params"]["mint_denom"]="'$DENOM'"
  | .app_state["mint"]["params"]["mint_denom"]="'$DENOM'"
  | .app_state["staking"]["params"]["bond_denom"]="'$DENOM'"
  | .app_state["consensus"]["params"]["abci"]["vote_extensions_enable_height"]="1"
  | .app_state["gov"]["params"]["expedited_voting_period"]="'$EXPEDITED_VOTING_PERIOD'"
  | .app_state["gov"]["params"]["voting_period"]="'$VOTING_PERIOD'"' \
    $n0cfgDir/genesis.json > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json


echo "--- Creating gentx..."
$NODE_BIN $home0 gentx $VAL0_KEY 1000000000$DENOM $kbt $cid
echo "--- Set POP to checkpointing module..."

$NODE_BIN $home0 gen-helpers ckpt-gen-key $VAL0_ADDR
# TODO, check how to use.
# $NODE_BIN $home0 create-genesis-bls
# $NODE_BIN $home0 add-genesis-bls $n0cfgDir/gen-bls-$VAL0_ADDR.json

$NODE_BIN $home0 collect-gentxs > /dev/null


if [[ -n "$BTC_BASE_HEADER_FILE" ]]; then
  jq '.app_state.btclightclient.btc_headers = [input]' $n0cfgDir/genesis.json $BTC_BASE_HEADER_FILE > $n0cfgDir/tmp_genesis.json
  mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json
fi

echo "--- Validating genesis..."
# initial_height bad sdk --" https://github.com/cosmos/cosmos-sdk/issues/18477
# $NODE_BIN $home0 validate-genesis

# Use perl for cross-platform compatibility
# Example usage: perl -i -pe 's/^param = ".*?"/param = "100"/' config.toml
echo "--- Modifying config..."
perl -i -pe 's|addr_book_strict = true|addr_book_strict = false|g' $n0cfg
perl -i -pe 's|external_address = ""|external_address = "tcp://127.0.0.1:26657"|g' $n0cfg
perl -i -pe 's|"tcp://127.0.0.1:26657"|"tcp://0.0.0.0:26657"|g' $n0cfg
perl -i -pe 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $n0cfg
perl -i -pe 's|log_level = "info"|log_level = "'$LOG_LEVEL'"|g' $n0cfg
perl -i -pe 's|timeout_commit = ".*?"|timeout_commit = "5s"|g' $n0cfg
perl -i -pe 's|cors_allowed_origins = \[\]|cors_allowed_origins = \["*"\]|g' $n0cfg

echo "--- Enabling node API and Swagger"
perl -i -pe 's|enable = false|enable = true|g'  $n0app
perl -i -pe 's|swagger = false|swagger = true|g'  $n0app

echo "--- Modifying app..."
perl -i -pe 's|minimum-gas-prices = ""|minimum-gas-prices = "0.05uquid"|g' $n0app
perl -i -pe 's|enable-unsafe-cors = false|enable-unsafe-cors = true|g' $n0app
perl -i -pe 's|enabled-unsafe-cors = false|enabled-unsafe-cors = true|g' $n0app

exit 0
