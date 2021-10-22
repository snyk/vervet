#!/usr/bin/env bash
set -eux
cd $(dirname $0)/..

if [ -z "${GOPATH:-}" ]; then
    tmp_gopath=$(mktemp -d)
    trap "chmod -R u+w $tmp_gopath; rm -rf $tmp_gopath" EXIT
    export GOPATH=$tmp_gopath
fi

rm -rf dist

NOW=$(date '+%Y%m%d%H%M')
COMMIT=$(git rev-parse --short HEAD)
export VERSION=$(git describe --abbrev=0)+${NOW}-${COMMIT}

mkdir -p ./dist/bin

for GOOS in linux darwin; do
    GOOS=$GOOS GOARCH=amd64 go build -a -o ./dist/bin/vervet-$GOOS-amd64 ./cmd/vervet
done
GOOS=windows GOARCH=amd64 go build -a -o ./dist/bin/vervet.exe ./cmd/vervet

cp packaging/npm/passthrough.js dist/bin/vervet
cp LICENSE ATTRIBUTIONS dist/

go get github.com/a8m/envsubst/cmd/envsubst
$GOPATH/bin/envsubst < packaging/npm/package.json.in > dist/package.json
