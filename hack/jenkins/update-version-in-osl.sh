#!/bin/bash -ex
cd "$(dirname "$0")" # cd to directory this file is in

VERSION=$(cat ../../release/VERSION)
NO_DOT_VERSION=$(echo $VERSION | sed -e 's/\.//g')
TODAY=$(date +%m%d%y)
sed -i '' -e "s/Wavefront Collector for Kubernetes \([0-9]\.\)*\([0-9]\)/Wavefront Collector for Kubernetes ${VERSION}/g" ../../open_source_licenses.txt
sed -i '' -e "s/WAVEFRONTKUBERNETESCOLLECTOR\([0-9]*\)\([a-zA-Z]*\)\([0-9]*\)/WAVEFRONTKUBERNETESCOLLECTOR${NO_DOT_VERSION}\2${TODAY}/g" ../../open_source_licenses.txt
