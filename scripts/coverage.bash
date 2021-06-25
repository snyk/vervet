#!/usr/bin/env bash
set -eux
cd $(dirname $0)/..

covfile=$(mktemp)
trap "rm -f $covfile" EXIT

go test ./... -count=1 -coverprofile=$covfile
go tool cover -html=$covfile
