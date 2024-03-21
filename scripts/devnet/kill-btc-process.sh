#!/bin/bash

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 || exit ; pwd -P )"
CHAIN_DIR="${CHAIN_DIR:-$CWD/node-data}"
BTC_HOME="$CHAIN_DIR/btc"

# btc_pid_file has both files together .-.
for btc_pid_file in $BTC_HOME/pid/*.pid; do
  echo PID looping ${btc_pid_file}
  if [ -f "$btc_pid_file" ]; then
    pid_value=$(cat "$btc_pid_file")
    if ps -p "$pid_value" > /dev/null; then
      kill -s 15 "$pid_value"
      echo -e "\t$pid_value killed"
    else
      echo -e "\tno process running"
    fi
  fi
done
