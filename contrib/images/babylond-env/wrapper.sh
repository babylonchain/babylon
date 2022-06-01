#!/usr/bin/env sh
set -euo pipefail
set -x

BINARY=/babylond/${BINARY:-babylond}
ID=${ID:-0}
LOG=${LOG:-babylond.log}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'babylond'"
	exit 1
fi

export BABYLONDHOME="/data/node${ID}/babylond"

if [ -d "$(dirname "${BABYLONDHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${BABYLONDHOME}" "$@" | tee "${BABYLONDHOME}/${LOG}"
else
  "${BINARY}" --home "${BABYLONDHOME}" "$@"
fi
