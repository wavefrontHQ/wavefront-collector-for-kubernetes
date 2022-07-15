#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)

function main() {
    cd "$REPO_ROOT"

    GOOS=linux go list -mod=readonly -deps -f '{{ if and (.DepOnly) (.Module) (not .Standard) }}{{ $mod := (or .Module.Replace .Module) }}{{ $mod.Path }}{{ end }}' ./... | grep -v github.com/wavefronthq/wavefront-collector-for-kubernetes | sort -u > /tmp/from_go_mod.txt
    grep '   >>> ' open_source_licenses.txt | grep -v Apache | grep -v Mozilla | awk '{print $2}' | rev | awk -F'v-' '{print $2}' | rev | sort -u > /tmp/from_open_source_licenses.txt

    diff -u /tmp/from_go_mod.txt /tmp/from_open_source_licenses.txt
}


main "$@"
