#!/bin/bash
set -o errexit -o nounset -o pipefail
command -v shellcheck >/dev/null && shellcheck "$0"

CONTRACT="babylon_contract"
OUTPUT_FOLDER="$(dirname "$0")/../bytecode"

echo "DEV-only: copy from local built instead of downloading"

cp -f  ../../../babylon-contract/artifacts/${CONTRACT}.wasm "$OUTPUT_FOLDER/"

cd ../../../babylon-contract
TAG=$(git rev-parse HEAD)
cd - 2>/dev/null
echo "$TAG" >"$OUTPUT_FOLDER/version.txt"
