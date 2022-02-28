#!/usr/bin/env bash
cd $(dirname $0)/../testdata
set -eu

go run ../cmd/vervet build
go run ../cmd/vervet generate -g generators.yaml

output=$(git status --porcelain) && [ -z "$output" ] || (
    echo "working directory not clean; testdata/output may be out of sync"
    exit 1
)
