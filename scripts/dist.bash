#!/usr/bin/env bash
set -eux
cd $(dirname $0)/..

# Require a version to be set
[ -n "${VERSION}" ]

# Ensure there is a GOPATH set
if [ -z "${GOPATH:-}" ]; then
    tmp_gopath=$(mktemp -d)
    trap "chmod -R u+w $tmp_gopath; rm -rf $tmp_gopath" EXIT
    export GOPATH=$tmp_gopath
fi
export PATH=$GOPATH/bin:$PATH

rm -rf dist

mkdir -p ./dist/bin

go generate ./internal/cmd/...

for GOOS in linux darwin; do
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=amd64 go build -a -o ./dist/bin/vervet-$GOOS-x64 ./cmd/vervet
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=arm64 go build -a -o ./dist/bin/vervet-$GOOS-arm64 ./cmd/vervet
done
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -o ./dist/bin/vervet.exe ./cmd/vervet

cp packaging/npm/passthrough.js dist/bin/vervet
cp README.md LICENSE ATTRIBUTIONS dist/

go install github.com/a8m/envsubst/cmd/envsubst@latest

envsubst < packaging/npm/package.json.in > dist/package.json
