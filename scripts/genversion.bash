#!/usr/bin/env bash
set -eux
cd $(dirname $0)/..

[ -n "${VERSION}" ]

cat << EOF > cmd/generate_version_init.go
// THIS IS A GENERATED FILE. DO NOT EDIT.

package cmd

func init() {
    Vervet.App.Version = "${VERSION}"
}
EOF
