#!/usr/bin/env sh
set -euo pipefail
set -x

DEBUG=${DEBUG:-0}
BINARY=/babylond/${BINARY:-babylond}
ID=${ID:-0}
LOG=${LOG:-babylond.log}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'babylond'"
	exit 1
fi

export BABYLONDHOME="/data/node${ID}/babylond"

if [ "$DEBUG" -eq 1 ]; then
  dlv --listen=:2345 --continue --headless=true --api-version=2 --accept-multiclient exec "${BINARY}" -- --home "${BABYLONDHOME}" "$@"
elif [ "$DEBUG" -eq 1 ] && [ -d "$(dirname "${BABYLONDHOME}"/"${LOG}")" ]; then
  dlv --listen=:2345 --continue --headless=true --api-version=2 --accept-multiclient exec "${BINARY}" -- --home "${BABYLONDHOME}" "$@" | tee "${BABYLONDHOME}/${LOG}"
elif [ -d "$(dirname "${BABYLONDHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${BABYLONDHOME}" "$@" | tee "${BABYLONDHOME}/${LOG}"
else
  "${BINARY}" --home "${BABYLONDHOME}" "$@"
fi
