#!/usr/bin/env bash
cd $(dirname $0)
set -eu

go run ./cmd/vervet compile -I ./testdata/resources/include.yaml ./testdata/resources/ ./testdata/output

output=$(git status --porcelain) && [ -z "$output" ] || (
    echo "working directory not clean; testdata/output may be out of sync"
    exit 1
)
