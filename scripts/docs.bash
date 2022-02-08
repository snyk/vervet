#!/usr/bin/env bash
set -eu
cd $(dirname $0)/..

# Ensure there is a GOPATH set
if [ -z "${GOPATH:-}" ]; then
    tmp_gopath=$(mktemp -d)
    trap "chmod -R u+w $tmp_gopath; rm -rf $tmp_gopath" EXIT
    export GOPATH=$tmp_gopath
fi
export PATH=$GOPATH/bin:$PATH

go install github.com/cmars/greenroom@v0.1.0

greenroom ./...

# Exit 0 if documentation unchanged, non-zero if there are changes.
# Supports CI gating on docs being up to date.

output=$(git status --porcelain)
if [ -z "$output" ]; then
    echo "No documentation changes"
    exit 0
fi

echo "Documentation has changed: $output"
exit 1
