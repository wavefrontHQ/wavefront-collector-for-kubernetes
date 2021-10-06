#!/usr/bin/env bash
set -ex

echo ${OLD_VERSION}
## Bump to next version

sed -i "s/${OLD_VERSION}/${NEXT_VERSION}/g" "deploy/kubernetes/5-collector-daemonset.yaml"
sed -i "s/${OLD_VERSION}/${NEXT_VERSION}/g" "deploy/openshift/collector/3-collector-deployment.yaml"
echo "$NEXT_VERSION" > ./VERSION

git commit -am "bump version to ${NEXT_VERSION}"

git push --set-upstream origin "$GIT_BRANCH"

gh pr create --base master --fill --head "$GIT_BRANCH" --web

curl -v \
  -X POST \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/wavefrontHQ/wavefront-collector-for-kubernetes/pulls \
  -d '{"head":"$GIT_BRANCH","base":"master"}'