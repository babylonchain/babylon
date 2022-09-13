#!/usr/bin/env sh
set -euo pipefail
#set -x

# btcctl will be looking for this file, but the wallet doesn't create it.
mkdir -p /root/.btcwallet && touch /root/.btcwallet/btcwallet.conf
mkdir -p /root/.btcd      && touch /root/.btcd/btcd.conf

# Create a wallet with and a miner address, then mine enough blocks for the miner to have some initial balance.

BITCOIN_CONF=${BITCOIN_CONF:-/bitcoinconf}
MINING_ADDR_FILE="${BITCOIN_CONF}/mining.addr"
CERT_FILE="${BITCOIN_CONF}/rpc.cert"
KEY_FILE="${BITCOIN_CONF}/rpc.key"
WALLET_CERT_FILE="${BITCOIN_CONF}/rpc-wallet.cert"
WALLET_KEY_FILE="${BITCOIN_CONF}/rpc-wallet.key"

echo "Creating certificates..."
gencerts -d $BITCOIN_CONF -H $CLIENT_HOST -f

echo "Starting btcd..."
btcd --simnet -u $RPC_USER -P $RPC_PASS --rpclisten=0.0.0.0:18556 --listen=0.0.0.0:18555 \
      --rpccert $CERT_FILE --rpckey $KEY_FILE 2>&1 &
BTCD_PID=$!

echo "Creating a wallet..."
# Used autoexpect to create the wallet in the first instance.
# https://stackoverflow.com/questions/4857702/how-to-provide-password-to-a-command-that-prompts-for-one-in-bash
expect btcwallet_create.exp $RPC_USER $RPC_PASS $WALLET_PASS $WALLET_CERT_FILE $WALLET_KEY_FILE $CERT_FILE

echo "Starting btcwallet..."
btcwallet --simnet -u $RPC_USER -P $RPC_PASS --rpclisten=0.0.0.0:18554 \
          --rpccert $WALLET_CERT_FILE --rpckey $WALLET_KEY_FILE --cafile $CERT_FILE 2>&1 &
BTCWALLET_PID=$!

# Allow some time for the wallet to start
sleep 5

echo "Creating miner address..."
MINING_ADDR=$(btcctl --simnet --wallet -u $RPC_USER -P $RPC_PASS --rpccert $WALLET_CERT_FILE getnewaddress)
echo $MINING_ADDR > $MINING_ADDR_FILE

echo "Restarting btcd with mining address $MINING_ADDR..."
kill -9 $BTCD_PID
sleep 1
btcd --simnet -u $RPC_USER -P $RPC_PASS --rpclisten=0.0.0.0:18556 --listen=0.0.0.0:18555 \
     --rpccert $CERT_FILE --rpckey $KEY_FILE --miningaddr=$MINING_ADDR 2>&1 &
BTCD_PID=$!

# Allow btcd to start
sleep 5

echo "Generating enough blocks for the first coinbase to mature..."
btcctl --simnet -u $RPC_USER -P $RPC_PASS --rpccert $CERT_FILE generate 100

# Allow some time for the wallet to catch up.
sleep 5

echo "Checking balance..."
btcctl --simnet --wallet -u $RPC_USER -P $RPC_PASS --rpccert $WALLET_CERT_FILE getbalance

echo "Generating a block every ${GENERATE_INTERVAL_SECS} seconds."
echo "Press [CTRL+C] to stop..."
while true
do
  btcctl --simnet -u $RPC_USER -P $RPC_PASS --rpccert $CERT_FILE generate 1

  sleep ${GENERATE_INTERVAL_SECS}
done
