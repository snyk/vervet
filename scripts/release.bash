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

go install github.com/goreleaser/goreleaser@latest

# Push tags
git tag ${VERSION}
git push --tags ${GIT_REMOTE:-origin}

# Publish npm package
(cd dist; npm publish)

# Github release
goreleaser release --rm-dist
