# Integration development:
This directory contains helpful scripts to aid in dashboard development.

Follow the below steps to develop an integration dashboard:

1. From the current path in collector repo, run the below script to create a development copy of the dashboard. `<dasboard-name>` should be the `url` name of the json file from [integration repo](https://github.com/sunnylabs/integrations/tree/master/kubernetes/dashboards) (For instance: `-d integration-kubernetes-control-plane`).
```
./scripts/start-dashboard-development.sh  -t $WAVEFRONT_TOKEN -d <dasboard-name>
```
Note: This would output a link to a dashboard in nimba. Remember to login to nimba and switch to `k8s-saas-team` before accessing the dashboard link.

2. PM or engineering team member iterate on dev dashboard returned from `start-dashboard-development.sh` (For instance: `kubernetes-control-plane-dev`)

3. When dev dashboard is ready, run the below script to push the changes to a branch in integration repo. `<dev-dashboard-name>` is the development dashboard's name created in step 1 (For instance: `-d kubernetes-control-plane-dev`). For branch names use the format `k8po/kubernetes<-any-details-as-seen-fit>`.
```
./scripts/merge-dashboard.sh  -t $WAVEFRONT_TOKEN -d <dev-dashboard-name> -b <integration-branch-name>
```
