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

# Publish npm package
if [ ! -e "dist/.npmrc" ]; then
    echo "//registry.npmjs.org/:_authToken=${NPM_TOKEN}" > dist/.npmrc
fi
(cd dist; npm publish)

# Github release
# Do this last; if it fails, it's easy to create a release in the UI.
# Pushing the tags and publishing to NPM are more important.
go install github.com/goreleaser/goreleaser@v1.6.3
GITHUB_TOKEN=${GH_TOKEN} goreleaser release --rm-dist
