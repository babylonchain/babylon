#!/usr/bin/env sh
set -euo pipefail
set -x

BINARY=/babylond/${BINARY:-babylond}
LOG=${LOG:-babylond.log}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'babylond'"
	exit 1
fi

export BABYLONDHOME="${HOME:-/data/node0/babylond}"

if [ -d "$(dirname "${BABYLONDHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${BABYLONDHOME}" "$@" 2>&1 | tee "${BABYLONDHOME}/${LOG}"
else
  "${BINARY}" --home "${BABYLONDHOME}" "$@"
fi
