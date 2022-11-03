#!/usr/bin/env bash
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "$here/git.sh"


if [ "${GITHUB_KEY+defined}" = defined ] && [ -n "$GITHUB_KEY" ]; then
  echo "Adding git key from GITHUB_KEY"
  source "$here/add-key.sh"
fi

echo "Running OSSPI"

INTERACTION_TYPE='Distributed - Calling Existing Classes'
declare -a close_out_package_managers_flag
close_out_package_managers_flag=("-a" "golang" "-a" "rubygem" "-a" "maven" "-a" "npm" "-a" "gradle" "-a" "bower" "-a" "other")

echo "$USERNAME" "$API_KEY" > apiKeyFile

declare -a scanning_params_flag
if [ "${OSSPI_SCANNING_PARAMS+defined}" = defined ] && [ -n "$OSSPI_SCANNING_PARAMS" ]; then
  printf "%s" "$OSSPI_SCANNING_PARAMS" > "/scanning-params.yaml"
  scanning_params_flag=("--conf" "/scanning-params.yaml")
  printf "Using configured OSSPI_SCANNING_PARAMS:\n%s\n\n" "$OSSPI_SCANNING_PARAMS"
else
  scanning_params_flag=("--conf" "scanning-params.yaml")
fi

declare -a ignore_package_flag
if [ "${OSSPI_IGNORE_RULES+defined}" = defined ] && [ -n "$OSSPI_IGNORE_RULES" ]; then
  printf "%s" "$OSSPI_IGNORE_RULES" > "/ignore-rules.yaml"
  ignore_package_flag=("--ignore-package-file" "/ignore-rules.yaml")
  printf "Using configured OSSPI_IGNORE_RULES:\n%s\n\n" "$OSSPI_IGNORE_RULES"
fi

repo_name=
if ! git_repo_name "$REPO" 'repo_name'; then
  echo "Error getting repo name" >&2
  return 1
fi
echo "Repo name: $repo_name"

declare -a package_group_name_flag
package_group_name_flag=("-gn" "$repo_name")

if [ "${OSM_PACKAGE_GROUP_NAME+defined}" = defined ] && [ -n "$OSM_PACKAGE_GROUP_NAME" ]; then
  echo "Using OSM_PACKAGE_GROUP_NAME: $OSM_PACKAGE_GROUP_NAME"
  package_group_name_flag=("-gn" "$OSM_PACKAGE_GROUP_NAME")
else
  echo "Using repo name as OSM package group name: $repo_name"
fi

repo_commit=
if ! git_repo_commit "$REPO" 'repo_commit'; then
  echo "Error getting repo commit" >&2
  return 1
fi
echo "Repo commit (and package group version): $repo_commit"
osm_package_group_version="$repo_commit"

pushd "$REPO"
  echo "$USERNAME" "$API_KEY" > apiKeyFile

  if [ "${PREPARE+defined}" = defined ] && [ -n "$PREPARE" ]; then
    printf "Running Prepare Command:\n%s\n\n" "$PREPARE"
    bash -c "$PREPARE"
  fi

  set -x

  osspi scan bom \
    "${scanning_params_flag[@]}" \
    "${ignore_package_flag[@]}" \
    --format json \
    --output-dir "$REPO"_bom
  
  osspi scan signature \
    "${scanning_params_flag[@]}" \
    "${ignore_package_flag[@]}" \
    --format json \
    --output-dir "$REPO"_signature
  
  # If nothing was found through bom scan, then results file is not created
  declare -a input_bom_result_flag
  if [ -f "$REPO"_bom/osspi_bom_detect_result.json ]; then
    input_bom_result_flag=('--input' "$REPO"_bom/osspi_bom_detect_result.json)
  fi

  osspi merge \
    "${input_bom_result_flag[@]}" \
    --input "$REPO"_signature/osspi_signature_detect_result.json \
    --output total_reports.yaml

  set +x

  str='[]'
  if [[ $(< total_reports.yaml) = "$str" ]]; then
    echo "Scan results are empty, exiting..."
    exit 0
  fi

  declare -a osstp_dry_run_flag
  if [ "${OSSTP_LOAD_DRY_RUN+defined}" = defined ] && [ "$OSSTP_LOAD_DRY_RUN" = 'true' ]; then
    osstp_dry_run_flag=('-n')
    echo "Dry run mode enabled for osstp-load"
  fi

  declare -a osstp_multiple_group_versions_flag
  if [ "${OSSTP_MULTIPLE_GROUP_VERSIONS+defined}" = defined ] && [ "$OSSTP_MULTIPLE_GROUP_VERSIONS" = 'true' ]; then
    osstp_multiple_group_versions_flag=('--multiple-group-versions')
    echo "Multiple group versions enabled for osstp-load"
  fi

  set -x
  
  osstp-load.py \
    "${osstp_dry_run_flag[@]}" \
    -S "$OSM_ENVIRONMENT" \
    -F \
    -A apiKeyFile \
    -I "$INTERACTION_TYPE" \
    -R "$PRODUCT"/"$VERSION" \
    --noinput \
    "${close_out_package_managers_flag[@]}" \
    "${package_group_name_flag[@]}" \
    -gv "$osm_package_group_version" \
    -gl 'norsk-to-osspi' \
    "${osstp_multiple_group_versions_flag[@]}" \
    total_reports.yaml
  
  set +x
popd
