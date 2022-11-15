# Integration development

Changes to a dashboard are typically done in the following ways.
- [Updating an existing dashboard that is already released](#updating-an-existing-dashboard-that-is-already-released)
- [Creating a new dashboard to iterate in a feature branch](#creating-a-new-dashboard-to-iterate-in-a-feature-branch)

## Updating an existing dashboard that is already released
1. Run the below script to create a development copy of the dashboard from nimba API.
    ```
    cd ~/workspace/wavefront-collector-for-kubernetes/hack/integrations/
    ./scripts/start-dashboard-development.sh  -t $WAVEFRONT_TOKEN -d <dasboard-url>
    ```
   * `<dasboard-url>` should be the `url` value of the json file from [integration repo](https://github.com/sunnylabs/integrations/tree/master/kubernetes/dashboards) (For instance: `-d integration-kubernetes-control-plane`).
   
   **Note:** The script would output a link to a dashboard in nimba. Remember to login to nimba and switch to `k8s-saas-team` before accessing the dashboard link.
2. PM or engineering team member iterate on dev dashboard returned from `start-dashboard-development.sh` (For instance: `kubernetes-control-plane-dev`)
3. Create a new branch in integrations repo and push it upstream. For branch names use the format `k8po/kubernetes<-any-details-as-seen-fit>`. If the development effort spans multiple jira stories, do not use jira story numbers in branch name.
4. Periodically pull the changes from the dev dashboard to the integration repo branch by running the below script.
    ```
    ~/workspace/wavefront-collector-for-kubernetes/hack/integrations/scripts/merge-dashboard.sh  -t $WAVEFRONT_TOKEN -d <dev-dashboard-name> -b <integration-branch-name>
    ```
   * `<dev-dashboard-name>` is the development dashboard's url created in step 2 (For instance: `-d kubernetes-control-plane-dev`). 
   * `<integration-branch-name>` is the branch name created in the integrations repo in the previous step.
5. Run the below dashboard validation script and fix any identified linting problems.
   ```
   ruby ~/workspace/integrations/tools/dashboards_validator.rb ~/workspace/integrations/kubernetes/dashboards/<dashboard-json-file-name>
   ```
6. Verify and commit the changes made to the dashboard in integrations repo.
7. Once the dev dashboard looks ready, follow the steps under [Merge the dashboard](https://confluence.eng.vmware.com/display/CNA/Technical+References#TechnicalReferences-Mergethedashboard).

## Creating a new dashboard to iterate in a feature branch

### Create a new dashboard json in development branch
1. Create a new branch in integrations repo and push it upstream. For branch names use the format `k8po/kubernetes<-any-details-as-seen-fit>`. If the development effort spans multiple jira stories, do not use jira story numbers in branch name.
2. Copy the json content of the [dashboard template](https://nimba.wavefront.com/u/5Ht7N57QKy?t=k8s-saas-team) following the [instructions here](https://docs.wavefront.com/ui_dashboards.html#edit-the-dashboard-json)
   and put it into a new file at `~/workspace/integrations/kubernetes/dashboards/integration-kubernetes-<new-dashboard>.json`
   
   **Note:** Remember to login to `nimba` and switch to `k8s-saas-team` before accessing the dashboard link.
3. Edit `integration-kubernetes-<new-dashboard>.json` as shown below
   ```
   "name": "Kubernetes <new-dashboard>",
   "url": "integration-kubernetes-<new-dashboard>",
   "systemDashboardVersion": 1,
   ```
4. Commit the new dashboard json into the branch created in integrations repo.

### Create a dev dashboard (ie., dashboard not in master) to iterate on

1. Clone a dev dashboard from the [dashboard template](https://nimba.wavefront.com/u/5Ht7N57QKy?t=k8s-saas-team) and name it following the below pattern
   ```
    Name: Kubernetes <new-dashboard>
    URL: kubernetes-<new-dashboard>-dev
   ```
2. Make changes to the dev dashboard (`kubernetes-<new-dashboard>-dev`).
3. Periodically pull the changes from the dev dashboard to the integration repo branch by running the below script.
    ```
    ~/workspace/wavefront-collector-for-kubernetes/hack/integrations/scripts/merge-dashboard.sh  -t $WAVEFRONT_TOKEN -d <dev-dashboard-name> -b <integration-branch-name>
    ```
    * `<dev-dashboard-name>` is the development dashboard's `URL` given in step 1 (For instance: `-d kubernetes-control-plane-dev`).
    * `<integration-branch-name>` is the branch name created in [Create a new dashboard in development branch](#create-a-new-dashboard-in-development-branch)
4. Run the below dashboard validation script and fix any identified linting problems.
    ```
    ruby ~/workspace/integrations/tools/dashboards_validator.rb ~/workspace/integrations/kubernetes/dashboards/<dashboard-json-file-name>
    ```
5. Verify and commit the changes made to the dashboard in integrations repo.
6. Once the dev dashboard looks ready, follow the steps under [Merge the dashboard](https://confluence.eng.vmware.com/display/CNA/Technical+References#TechnicalReferences-Mergethedashboard).
