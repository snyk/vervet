#!/usr/bin/env bash
set -euo pipefail
cd $(dirname $0)/..

mkdir -p ./dist/bin

for GOOS in linux darwin; do
    GOOS=$GOOS GOARCH=amd64 go build -a -o ./dist/bin/vervet-$GOOS-amd64 ./cmd/vervet
done
GOOS=windows GOARCH=amd64 go build -a -o ./dist/bin/vervet.exe ./cmd/vervet

