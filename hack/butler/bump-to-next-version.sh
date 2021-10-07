#!/usr/bin/env bash
set -ex

echo "Bumping ${OLD_VERSION} to ${NEXT_VERSION} in deploy/kubernetes/5-collector-daemonset.yaml"
sed -i "s/${OLD_VERSION}/${NEXT_VERSION}/g" "deploy/kubernetes/5-collector-daemonset.yaml"
echo "$NEXT_VERSION" >./release/VERSION

git commit -am "bump version to ${NEXT_VERSION}"

git push --set-upstream origin "${GIT_BUMP_BRANCH_NAME}"

curl \
  -X POST \
  -H "Authorization: token ${TOKEN}" \
  -d "{\"head\":\"${GIT_BUMP_BRANCH_NAME}\",\"base\":\"master\",\"title\":\"Bump version to ${NEXT_VERSION}\"}" \
  https://api.github.com/repos/wavefrontHQ/wavefront-collector-for-kubernetes/pulls
