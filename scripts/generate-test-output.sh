#!/usr/bin/env bash
cd $(dirname $0)/../testdata
set -eu

go run ../cmd/vervet compile "$@"
go run ../cmd/vervet version new --force --version 2021-09-01 --stability beta testdata newthing "$@"

output=$(git status --porcelain) && [ -z "$output" ] || (
    echo "working directory not clean; testdata/output may be out of sync"
    exit 1
)
