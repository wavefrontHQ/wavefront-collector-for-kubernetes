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

    local PREV_COMMIT=
    local CURR_COMMIT=

    while getopts ":p:c:" opt; do
        case $opt in
        p)
            PREV_COMMIT="$OPTARG"
            ;;
        c)
            CURR_COMMIT="$OPTARG"
            ;;
        \?)
            print_usage_and_exit "Invalid option: -$OPTARG"
            ;;
        esac
    done

    if [[ -z $PREV_COMMIT ]]; then
        print_msg_and_exit "previous commit (-p) required"
    fi

    if [[ -z $CURR_COMMIT ]]; then
        print_msg_and_exit "current commit (-c) required"
    fi

    if ! git diff --name-only "$PREV_COMMIT...$CURR_COMMIT" | grep -c go.sum > /dev/null; then
        echo "no go dependencies have changed"
        exit 0
    fi

    GOOS=darwin go list -mod=readonly -deps -f '{{ if and (.DepOnly) (.Module) (not .Standard) }}{{ $mod := (or .Module.Replace .Module) }}{{ $mod.Path }}{{ end }}' ./... | sort -u  | grep -v github.com/wavefronthq/wavefront-collector-for-kubernetes > /tmp/from_go_mod.txt
    grep '   >>> ' open_source_licenses.txt | sort -u | grep -v Apache | grep -v Mozilla | awk '{print $2}' | rev | awk -F'v-' '{print $2}' | rev > /tmp/from_open_source_licenses.txt

    diff -u /tmp/from_go_mod.txt /tmp/from_open_source_licenses.txt
}

main "$@"
