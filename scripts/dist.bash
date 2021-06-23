#!/usr/bin/env bash
set -eux
cd $(dirname $0)/..

# TODO: get version from latest release tag, append short hash
export VERSION=0.0.1

mkdir -p ./dist/bin

for GOOS in linux darwin; do
    GOOS=$GOOS GOARCH=amd64 go build -a -o ./dist/bin/vervet-$GOOS-amd64 ./cmd/vervet
done
GOOS=windows GOARCH=amd64 go build -a -o ./dist/bin/vervet.exe ./cmd/vervet

cp packaging/npm/passthrough.js dist/bin/vervet
envsubst < packaging/npm/package.json.in > dist/package.json

