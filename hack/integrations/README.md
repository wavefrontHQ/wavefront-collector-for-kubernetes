# Dashboard development

1. Run the below script to create a development copy of the dashboard from nimba API.
    ```
    ~/workspace/wavefront-collector-for-kubernetes/hack/integrations/scripts/create-or-update-dashboard-in-ui.sh  -t $WAVEFRONT_TOKEN -d <DASHBOARD_TO_CLONE> -n <NEW_DASHBOARD>
    ```
   * `<DASHBOARD_TO_CLONE>` should be the url slug for the dashboard you want to clone. For instance `-d integration-kubernetes-clusters` for https://nimba.wavefront.com/dashboards/integration-kubernetes-clusters.
     If it is not set, it will default to `integration-dashboard-template`
   * `<NEW_DASHBOARD>` is the new dashboard in UI to create (For instance: `-n kubernetes-K8SSAAS-123`).

   **Note:** The script would output a link to a dashboard in nimba. Remember to login to nimba and switch to `k8s-saas-team` before accessing the dashboard link.
2. PM or engineering team member iterate on dev dashboard created by `create-or-update-dashboard-in-ui.sh` (For instance: `kubernetes-K8SSAAS-123`).
3. Periodically pull and validate the changes from the dev dashboard to the integration repo branch by running the below script.
    ```
    ~/workspace/wavefront-collector-for-kubernetes/hack/integrations/scripts/merge-dashboard.sh  -t $WAVEFRONT_TOKEN -s <SOURCE_DASHBOARD> -d <DEST_DASHBOARD> -b <BRANCH_NAME>
    ```
   * `<SOURCE_DASHBOARD>` is the source dashboard's url in UI that you want to copy **from** (For instance: `-s kubernetes-K8SSAAS-123`). 
   * `<DEST_DASHBOARD>` is the destination dashboard's url in integrations repo to copy **to** (For instance: `-d integration-kubernetes-clusters`). 
   * `<BRANCH_NAME>` is the branch name to create or switch to in the integrations repo (For instance: `-b k8po/new-dashboard-work`).
4. Fix any validation issues identified in previous step. Verify, commit and push the local changes made to the dashboard in integrations repo.
5. Once the dev dashboard looks ready, follow the steps under [Merge the dashboard](https://confluence.eng.vmware.com/display/CNA/Technical+References#TechnicalReferences-Mergethedashboard).
