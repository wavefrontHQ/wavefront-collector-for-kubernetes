#!/usr/bin/env bash
set -euo pipefail

declare -a scanning_params_flag
if [ "${OSSPI_SCANNING_PARAMS+defined}" = defined ] && [ -n "$OSSPI_SCANNING_PARAMS" ]; then
  printf "%s" "$OSSPI_SCANNING_PARAMS" > scanning-params.yaml
  scanning_params_flag=("--conf" "scanning-params.yaml")
else
  scanning_params_flag=("--conf" "scanning-params.yaml")
fi

declare -a ignore_package_flag
if [ "${OSSPI_IGNORE_RULES+defined}" = defined ] && [ -n "$OSSPI_IGNORE_RULES" ]; then
  printf "%s" "$OSSPI_IGNORE_RULES" > ignore-rules.yaml
  ignore_package_flag=("--ignore-package-file" "ignore-rules.yaml")
fi

if [ "${PREPARE+defined}" = defined ] && [ -n "$PREPARE" ]; then
  bash -c "$PREPARE" > /dev/null 2>&1
fi

set -x

$HOME/.osspicli/osspi/osspi scan bom \
  "${scanning_params_flag[@]}" \
  "${ignore_package_flag[@]}" \
  --format json \
  --output-dir "$REPO"_bom > /dev/null 2>&1

$HOME/.osspicli/osspi/osspi scan signature \
  "${scanning_params_flag[@]}" \
  "${ignore_package_flag[@]}" \
  --format json \
  --output-dir "$REPO"_signature > /dev/null 2>&1

# If nothing was found through bom scan, then results file is not created
declare -a input_bom_result_flag
if [ -f "$REPO"_bom/osspi_bom_detect_result.json ]; then
  input_bom_result_flag=('--input' "$REPO"_bom/osspi_bom_detect_result.json)
fi

$HOME/.osspicli/osspi/osspi merge \
  "${input_bom_result_flag[@]}" \
  --input "$REPO"_signature/osspi_signature_detect_result.json \
  --output "$OUTPUT" > /dev/null 2>&1

set +x