#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$REPO_ROOT/hack/test/deploy/k8s-utils.sh"

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 -p <commit> -c <commit>"
  echo -e "\t-p previous commit (required)"
  echo -e "\t-c current commit (required)"
  echo "returns with exit code 0 when go.sum has not changed and exit code 1 when it has"
  exit 1
}

function main() {
    cd "$REPO_ROOT"

    GOOS=darwin go list -mod=readonly -deps -f '{{ if and (.DepOnly) (.Module) (not .Standard) }}{{ $mod := (or .Module.Replace .Module) }}{{ $mod.Path }}{{ end }}' ./... | sort -u  | grep -v github.com/wavefronthq/wavefront-collector-for-kubernetes > /tmp/from_go_mod.txt
    grep '   >>> ' open_source_licenses.txt | sort -u | grep -v Apache | grep -v Mozilla | awk '{print $2}' | rev | awk -F'v-' '{print $2}' | rev > /tmp/from_open_source_licenses.txt

    diff -u /tmp/from_go_mod.txt /tmp/from_open_source_licenses.txt
}

main "$@"
