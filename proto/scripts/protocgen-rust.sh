#!/usr/bin/env bash

set -eo pipefail

cd proto
proto_dirs=$(find ./babylon -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

buf mod update
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    basefile=$(basename $file)
    basefile_noext=${basefile%.*}
    if echo "$dir" | grep -q $basefile_noext; then
      echo "generate protobuf file for $file..."
      buf generate --template buf.gen.rust.yaml $file
    fi
  done
done
cd ..
