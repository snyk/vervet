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

# Push tags in CI environment
if [ -z $(git config user.email) ]; then
    git config credential.helper 'cache --timeout=120'
    git config user.email "vervet-ci@noreply.snyk.io"
    git config user.name "Vervet CI"
fi
git tag ${VERSION}
git push -q https://${GH_TOKEN}@github.com/snyk/vervet.git --tags

# Publish npm package
(cd dist; npm publish)

# Github release
# Do this last; if it fails, it's easy to create a release in the UI.
# Pushing the tags and publishing to NPM are more important.
go install github.com/goreleaser/goreleaser@latest
GITHUB_TOKEN=${GH_TOKEN} goreleaser release --rm-dist
