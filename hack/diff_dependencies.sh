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
  pwd
  echo "./$SCRIPT_DIR/../osspi/tasks/osspi/"
  . ./$SCRIPT_DIR/../osspi/tasks/osspi/scan-source.sh


  GOOS=linux go list -mod=readonly -deps -f '{{ if and (.DepOnly) (.Module) (not .Standard) }}{{ $mod := (or .Module.Replace .Module) }}{{ $mod.Path }}{{ end }}' ./... | grep -v $REPO | sort -u > $TEMP_DIR/from_go_mod.txt
  grep '   >>> ' open_source_licenses.txt | grep -v Apache | grep -v Mozilla | awk '{print $2}' | rev | awk -F'v-' '{print $2}' | rev | sort -u > $TEMP_DIR/from_open_source_licenses.txt

#  TODO: Until we can figure out how to find dependency change in docker images,
#  only consider new dependency additions to go mod file.
#  diff -u $TEMP_DIR/from_go_mod.txt $TEMP_DIR/from_open_source_licenses.txt

  NEW_GO_DEPENDENCIES=$(comm -13 <(sort $TEMP_DIR/from_open_source_licenses.txt | uniq) <(sort $TEMP_DIR/from_go_mod.txt | uniq))
  NEW_GO_DEPENDENCIES_COUNT="$(printf "%s" "${NEW_GO_DEPENDENCIES//[!$'\n']/}" | grep -c '^')"
  if [[ $NEW_GO_DEPENDENCIES_COUNT -ne 0 ]]; then
    echo "Found $NEW_GO_DEPENDENCIES_COUNT new dependencies in go.mod that are not in open_source_licenses.txt:"
    printf "%s\n" $NEW_GO_DEPENDENCIES
    exit 1
  fi
}


main "$@"
