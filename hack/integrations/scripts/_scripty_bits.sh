#   local INTEGRATIONS_REPO="$HOME/workspace/integrations"
#   local BRANCH_NAME="k8po/kubernetes-${BRANCH_NAME_SUFFIX}"
  # Change the url field to match the integration url instead of the dev dashboard url
#  local DASHBOARD_URL="integration-$(echo "${DASHBOARD_DEV_URL}" | sed 's/-dev//')"
#  jq ".url = \"${DASHBOARD_URL}\"" ${DASHBOARD_DEV_URL}.json > ${DASHBOARD_URL}.json
#
#  # Copy dashboard version from integration feature branch and increment it
#  local VERSION=$(($(jq ".systemDashboardVersion" ${INTEGRATION_DIR}/kubernetes/dashboards/${DASHBOARD_URL}.json)+1))
#  jq ". += {"systemDashboardVersion":${VERSION}}" ${DASHBOARD_URL}.json > "tmp" && mv "tmp" ${DASHBOARD_URL}.json

#  pushd_check "$INTEGRATIONS_REPO"
#    git stash
#    git checkout master
#    git pull
#    git checkout -b "${BRANCH_NAME}"
#    git push --set-upstream origin "${BRANCH_NAME}"
#
#    cat ${DASHBOARD_URL}.json > ${INTEGRATION_DIR}/kubernetes/dashboards/${DASHBOARD_URL}.json
#    echo Check your integration repo for changes.
#  popd_check "$INTEGRATIONS_REPO"



# get story about creating dashboard
start-dashboard-development.sh # -create from template
create-or-update-wavefront-dashboard.sh
# work in UI


# about to leave for lunch or for the day and want to commit changes
get-dashboard.sh -c nimba -t $WAVEFRONT_TOKEN -d test-put-verb \
  -o local-dashboard-copy.json
sort-dashboard.sh # ...
# git stuff
cp local-dashboard-copy.json INTEGRATION_DIR/...

download-and-sort-and-copy-dashboard-to-integrations.sh

merge-dashboard-and-update-version.sh -i local-dashboard-copy.json \
  -o ~/workspace/integrations/kubernetes/dashboards/integration-test-put-verb.json
# sort-dashboard.sh # fix the dashboard for easy viewing of changes, clean up after self
# manually review changes
# commit + push



# come back from lunch or next day
update-dashboard-from-repo.sh
update-wavefront-dashboard.sh

# merge-dashboard should be responsible both for
