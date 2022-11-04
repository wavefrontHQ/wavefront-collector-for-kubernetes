#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 [flags] [options]"
  echo -e "\t-r repository name (required)"
  echo "Run this script from the repository where you want to compare the go.mod file with the open_source_licenses.txt file."
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
  SCRIPT_DIR=$(dirname "$0")

  if [[ ! -f "$HOME/.osspicli/osspi/osspi" ]]; then
    echo "installing osspi..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
      bash -c "$(curl -fsSL https://build-artifactory.eng.vmware.com/osspicli-local/beta/osspicli-darwin/install.sh)"
    else
      bash -c "$(curl -fsSL https://build-artifactory.eng.vmware.com/osspicli-local/beta/osspicli/install.sh)"
    fi
    echo "successfully installed osspi: $($HOME/.osspicli/osspi/osspi --version)"
  else
    echo "osspi already installed: $($HOME/.osspicli/osspi/osspi --version)"
  fi

  OSSPI_SCANNING_PARAMS=$(cat <<EOF
  enable: true
  include_bomtools: "go_mod"
  search_depth: 5

  # exclude for signature scans
  exclude_patterns:
    - vendor
EOF
  )
  echo "OSSPI_SCANNING_PARAMS: $OSSPI_SCANNING_PARAMS"

  OSSPI_IGNORE_RULES=$(cat <<EOF
  - name_regex: onsi\/ginkgo
    version_regex: .*
  - name_regex: gomega
    version_regex: .*
EOF
  )
  echo "OSSPI_IGNORE_RULES: $OSSPI_IGNORE_RULES"

  PREPARE="go mod vendor"
  echo "PREPARE: $PREPARE"

  OUTPUT="scan-report.json"

#
  . ./$SCRIPT_DIR/../osspi/tasks/osspi/scan-source.sh


#  GOOS=linux go list -mod=readonly -deps -f '{{ if and (.DepOnly) (.Module) (not .Standard) }}{{ $mod := (or .Module.Replace .Module) }}{{ $mod.Path }}{{ end }}' ./... | grep -v $REPO | sort -u > $TEMP_DIR/from_go_mod.txt
  grep '   >>> ' open_source_licenses.txt | grep -v Apache | grep -v Mozilla | awk '{print $2}' | rev | awk -F'v-' '{print $2}' | rev | sort -u > $TEMP_DIR/from_open_source_licenses.txt
  cat scan-report.json | jq '.packages' | jq '.[] | {name} | add' | cut -d '"' -f2 | sort -u > $TEMP_DIR/from_osspi_scan.txt

  echo "Found new dependencies from osspi scan that are not in open_source_licenses.txt:"
  diff -u $TEMP_DIR/from_osspi_scan.txt $TEMP_DIR/from_open_source_licenses.txt | grep "^-[a-zA-Z]"
  echo "Found old dependencies in open_source_licenses.txt that are not in osspi scan:"
  diff -u $TEMP_DIR/from_osspi_scan.txt $TEMP_DIR/from_open_source_licenses.txt | grep "^+[a-zA-Z]"


#  ADDED_DEP=${diff -u $TEMP_DIR/from_osspi_scan.txt $TEMP_DIR/from_open_source_licenses.txt | grep "^-[a-zA-Z]"}
#  REMOVED_DEP=${diff -u $TEMP_DIR/from_osspi_scan.txt $TEMP_DIR/from_open_source_licenses.txt | grep "^+[a-zA-Z]"}
#
#  ADDED_DEP_COUNT="$(printf "%s" "${ADDED_DEP//[!$'\n']/}" | grep -c '^')"
#  if [[ $ADDED_DEP_COUNT -ne 0 ]]; then
#    echo "Found $NEW_GO_DEPENDENCIES_COUNT new dependencies from osspi scan that are not in open_source_licenses.txt:"
#    printf "%s\n" $ADDED_DEP
#    exit 1
#  fi
#  REMOVED_DEP_COUNT="$(printf "%s" "${REMOVED_DEP//[!$'\n']/}" | grep -c '^')"
#  if [[ $REMOVED_DEP_COUNT -ne 0 ]]; then
#    echo "Found $REMOVED_DEP_COUNT old dependencies in open_source_licenses.txt that are not in osspi scan:"
#    printf "%s\n" $REMOVED_DEP
#    exit 1
#  fi
}


main "$@"
