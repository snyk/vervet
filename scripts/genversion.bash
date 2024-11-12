#!/usr/bin/env bash
set -eux
cd $(dirname $0)/..

[ -n "${VERSION}" ]

sed -i -r "s/^(const cmdVersion = )\".*\"$/\1\"${VERSION}\"/" internal/cmd/cmd.go
