# Installation and configuration on OpenShift Origin 3.9

1. Login into Openshift master node.
2. Login into Openshift cluster `oc login -u <ADMIN_USER>`
3. Clone this Repo.
4. Create `wavefront-collector` namespace:
```
cd deploy/openshift
oc create -f 0-collector-namespace.yaml
```
Note: Step 5 and 6 are needed if you are planning to use Wavefront Proxy else go to step 7.

5. Login into Openshift web console and create storage under `wavefront-collector`. Select Access Mode as `RWX` and Size as `5 GiB`, give name to the storage and make a note of it.

6. Replace YOUR_CLUSTER, YOUR_API_TOKEN and STORAGE_NAME in `1-wavefront-proxy.yaml` and run:
```
oc create -f 1-wavefront-proxy.yaml
```
7. Deploy kube-state-metrics into your cluster:
```
oc create -f 2-kube-state.yaml
```
8. Edit the `wavefront` sink and `cluster name` in `collector/3-collector-deployment.yaml` based on the selected metric ingestion approach as given below.
#### Using Wavefront Proxy

```
--sink=wavefront:?proxyAddress=wavefront-proxy.wavefront-collector.svc.cluster.local:2878&clusterName=openshift-cluster&includeLabels=true
```

#### Using Direct Ingestion
```
--sink=wavefront:?server=https://<YOUR_INSTANCE>.wavefront.com&token=<YOUR_TOKEN>&clusterName=openshift-cluster&includeLabels=true
```
9. Deploy the Wavefront Collector
```
oc create -f collector
```
