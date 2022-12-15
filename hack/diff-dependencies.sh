#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 [flags] [options]"
  echo -e "\t-r repository name (required)"
  echo "Run this script from the repository where you want to compare its osspi scan result with the open_source_licenses.txt file."
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
  TEMP_DIR="$REPO"_compare
  mkdir $TEMP_DIR
  echo "TEMP_DIR: $TEMP_DIR"
  SCRIPT_DIR=$(dirname "$0")

  OSSPI_SCANNING_PARAMS=$(cat <<EOF
  enable: true
  include_bomtools: "go_mod"
  search_depth: 5
  # exclude for signature scans
  exclude_patterns:
    - vendor
EOF
  )
  OSSPI_IGNORE_RULES=$(cat <<EOF
  - name_regex: onsi\/ginkgo
    version_regex: .*
  - name_regex: gomega
    version_regex: .*
EOF
  )

  declare -a scanning_params_flag
  if [ "${OSSPI_SCANNING_PARAMS+defined}" = defined ] && [ -n "$OSSPI_SCANNING_PARAMS" ]; then
    printf "%s" "$OSSPI_SCANNING_PARAMS" >scanning-params.yaml
    scanning_params_flag=("--conf" "scanning-params.yaml")
  else
    scanning_params_flag=("--conf" "scanning-params.yaml")
  fi

  declare -a ignore_package_flag
  if [ "${OSSPI_IGNORE_RULES+defined}" = defined ] && [ -n "$OSSPI_IGNORE_RULES" ]; then
    printf "%s" "$OSSPI_IGNORE_RULES" >ignore-rules.yaml
    ignore_package_flag=("--ignore-package-file" "ignore-rules.yaml")
  fi

  PREPARE="go mod vendor"
  OUTPUT="scan-report.json"
  rm -rf "$OUTPUT"
  if [ "${PREPARE+defined}" = defined ] && [ -n "$PREPARE" ]; then
    bash -c "$PREPARE" >/dev/null 2>&1
  fi

  set -x

  $HOME/.osspicli/osspi/osspi scan bom \
    "${scanning_params_flag[@]}" \
    "${ignore_package_flag[@]}" \
    --format json \
    --output-dir "$REPO"_bom >/dev/null 2>&1

  $HOME/.osspicli/osspi/osspi scan signature \
    "${scanning_params_flag[@]}" \
    "${ignore_package_flag[@]}" \
    --format json \
    --output-dir "$REPO"_signature >/dev/null 2>&1

  # If nothing was found through bom scan, then results file is not created
  declare -a input_bom_result_flag
  RESULT_FILE="${REPO}_bom/osspi_bom_detect_result.json"
  if [[ -f ${RESULT_FILE} ]]; then
    input_bom_result_flag=('--input' "$REPO"_bom/osspi_bom_detect_result.json)
  fi

  $HOME/.osspicli/osspi/osspi merge \
    "${input_bom_result_flag[@]}" \
    --input "$REPO"_signature/osspi_signature_detect_result.json \
    --output "$OUTPUT" >/dev/null 2>&1

  grep '   >>> ' open_source_licenses.txt | grep -v Apache | grep -v Mozilla | awk '{print $2}' | rev | awk -F'v-' '{print $2}' | rev | sort -u > $TEMP_DIR/from_open_source_licenses.txt
  cat scan-report.json | jq '.packages' | jq '.[] | {name} | add' | cut -d '"' -f2 | sort -u > $TEMP_DIR/from_osspi_scan.txt

  EXIT_CODE=0
  ADDED_DEP=$(comm -13 <(sort $TEMP_DIR/from_open_source_licenses.txt | uniq) <(sort $TEMP_DIR/from_osspi_scan.txt | uniq))
  REMOVED_DEP=$(comm -13 <(sort $TEMP_DIR/from_osspi_scan.txt | uniq) <(sort $TEMP_DIR/from_open_source_licenses.txt | uniq))

  ADDED_DEP_COUNT="$(printf "%s" "${ADDED_DEP//[!$'\n']/}" | grep -c '^')"
  if [[ $ADDED_DEP_COUNT -ne 0 ]]; then
    echo "Found $ADDED_DEP_COUNT new dependencies from osspi scan that are not in open_source_licenses.txt:"
    printf "%s\n" $ADDED_DEP
    EXIT_CODE=8
  fi
  REMOVED_DEP_COUNT="$(printf "%s" "${REMOVED_DEP//[!$'\n']/}" | grep -c '^')"
  if [[ $REMOVED_DEP_COUNT -ne 0 ]]; then
    echo "Found $REMOVED_DEP_COUNT old dependencies in open_source_licenses.txt that are not in osspi scan:"
    printf "%s\n" $REMOVED_DEP
    EXIT_CODE=8
  fi
  exit "$EXIT_CODE"
}


main "$@"
