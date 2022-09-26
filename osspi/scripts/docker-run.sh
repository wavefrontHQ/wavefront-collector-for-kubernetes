# shellcheck shell=bash

set -euo pipefail

#if [ -f "$HOME/Downloads/my-git-key" ]; then
#  when golang project pulls in other private golang repos
#  GITHUB_KEY=$(<"$HOME/Downloads/my-git-key")
#  export GITHUB_KEY
#fi

IMAGE=projects.registry.vmware.com/tanzu_observability/kubernetes-operator-fluentd
export IMAGE

TAG=1.0.3-1.15.2
export TAG

API_KEY='<your OSM API key>'
export API_KEY

OSM_ENVIRONMENT='beta'
export OSM_ENVIRONMENT

OSSTP_LOAD_DRY_RUN=true
export OSSTP_LOAD_DRY_RUN

ENDPOINT='https://osm-beta.eng.vmware.com/'
export ENDPOINT

USERNAME='<your OSM email login'
export USERNAME

PRODUCT='Wavefront_K8_Operator'
export PRODUCT

VERSION='Latest'
export VERSION

# for ignoring specific packages after the scan
OSSPI_IGNORE_RULES=
export OSSPI_IGNORE_RULES

APPEND=true
export APPEND

CT_TRACKER_OS=debian
export CT_TRACKER_OS

docker run \
  -v ~/workspace/:/workspace/ \
  --env PREPARE \
  --env OSSPI_SCANNING_PARAMS \
  --env BLOB_SOURCES_CONFIG \
  --env IMAGE \
  --env TAG \
  --env API_KEY \
  --env OSM_ENVIRONMENT \
  --env OSSTP_LOAD_DRY_RUN \
  --env ENDPOINT \
  --env USERNAME \
  --env PRODUCT \
  --env VERSION \
  --env OSSPI_IGNORE_RULES \
  --env APPEND \
  --env CT_TRACKER_OS \
  -it harbor-repo.vmware.com/source_insight_tooling/osspi-runner:latest
