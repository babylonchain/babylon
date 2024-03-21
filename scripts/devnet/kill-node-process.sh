#!/bin/bash

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 || exit ; pwd -P )"

CHAIN_ID="${CHAIN_ID:-test-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
HOME_DIR="$CHAIN_DIR/$CHAIN_ID"

for dirnode in "$HOME_DIR"/n*; do
  echo Node DIR looping ${dirnode}
  if [ -d "$dirnode" ]
  then
    echo "$CHAIN_ID/$(basename "${dirnode}")"
    pid_file="$dirnode/pid"
    if [ -f "$pid_file" ]; then
      pid_value=$(cat "$pid_file")
      if ps -p "$pid_value" > /dev/null; then
        kill -s 15 "$pid_value"
        echo -e "\t$pid_value killed"
      else
        echo -e "\tno process running"
      fi
    fi
  fi
done
