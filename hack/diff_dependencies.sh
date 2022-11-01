#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
REPO=github.com/wavefronthq/wavefront-operator-for-kubernetes
TEMP_DIR=$(mktemp -d)

function main() {
    cd "$REPO_ROOT"
    echo $TEMP_DIR
    GOOS=linux go list -mod=readonly -deps -f '{{ if and (.DepOnly) (.Module) (not .Standard) }}{{ $mod := (or .Module.Replace .Module) }}{{ $mod.Path }}{{ end }}' ./... | grep -v $REPO | sort -u > $TEMP_DIR/from_go_mod.txt
    grep '   >>> ' open_source_licenses.txt | grep -v Apache | grep -v Mozilla | awk '{print $2}' | rev | awk -F'v-' '{print $2}' | rev | sort -u > $TEMP_DIR/from_open_source_licenses.txt
    diff -u $TEMP_DIR/from_go_mod.txt $TEMP_DIR/from_open_source_licenses.txt
}


main "$@"
