#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 [flags] [options]"
  echo -e "\t-r repository name (required)"
  exit 1
}

function main() {
  while getopts ":r:" opt; do
    case $opt in
    r)
      REPO="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
    esac
  done

  if [[ -z ${REPO} ]]; then
    print_usage_and_exit "repository name required"
  fi
  cd "$REPO_ROOT"
  TEMP_DIR=$(mktemp -d)

  echo "TEMP dir: $TEMP_DIR, REPO_ROOT: $REPO_ROOT, REPO: $REPO, current dir: "
  pwd
  GOOS=linux go list -mod=readonly -deps -f '{{ if and (.DepOnly) (.Module) (not .Standard) }}{{ $mod := (or .Module.Replace .Module) }}{{ $mod.Path }}{{ end }}' ./... | grep -v $REPO | sort -u > $TEMP_DIR/from_go_mod.txt
  grep '   >>> ' open_source_licenses.txt | grep -v Apache | grep -v Mozilla | awk '{print $2}' | rev | awk -F'v-' '{print $2}' | rev | sort -u > $TEMP_DIR/from_open_source_licenses.txt
  diff -u $TEMP_DIR/from_go_mod.txt $TEMP_DIR/from_open_source_licenses.txt
}


main "$@"
