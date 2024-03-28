#!/bin/bash

# USAGE:
# PATH_OF_PIDS=path/to/pids/*.pid ./kill-process.sh

# Kill all the process stored in the PATH_OF_PIDS files

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 || exit ; pwd -P )"
CHAIN_DIR="${CHAIN_DIR:-$CWD/data}"
CHAIN_ID="${CHAIN_ID:-test-1}"
PATH_OF_PIDS="${PATH_OF_PIDS:-$CHAIN_DIR/$CHAIN_ID/*.pid}"

for pid_file in $PATH_OF_PIDS; do
  echo PID looping ${pid_file}
  if [ -f "$pid_file" ]; then
    pid_value=$(cat "$pid_file")
    if ps -p "$pid_value" > /dev/null; then
      kill -s 15 "$pid_value"
      echo -e "\t$pid_value killed"
    else
      echo -e "\tno process running"
    fi
  fi
done
