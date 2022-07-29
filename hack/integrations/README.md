# Integration development:
This directory contains helpful scripts to aid in dashboard development.

Follow the below steps to develop an integration dashboard:

1. Create a development copy of the dashboard. Dashboard name should be the name of the json file from [integration repo](https://github.com/sunnylabs/integrations/tree/master/kubernetes/dashboards) (For instance: `-d integration-kubernetes-control-plane`).
```
./scripts/start-dashboard-development.sh  -t $WAVEFRONT_TOKEN -d <dasboard-name>
```
2. PM or engg member iterate on dev dashboard returned from `start-dashboard-development.sh` (For instance: `kubernetes-control-plane-dev`)


3. When dev dashboard is ready, push the changes to a branch in integration repo. `<dev-dashboard-name>` is the development dashboard's name created in step 1 (For instance: `-d kubernetes-control-plane-dev`)
```
./scripts/merge-dashboard.sh  -t $WAVEFRONT_TOKEN -d <dev-dashboard-name> -b <integration-branch-name>
```