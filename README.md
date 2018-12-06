# wavefront-kubernetes-collector [![build status][ci-img]][ci] [![Go Report Card][go-report-img]][go-report] [![Docker Pulls][docker-pull-img]][docker-img]

This collector enables monitoring Kubernetes clusters and sending metrics to [Wavefront](https://www.wavefront.com).

The collector scrapes the Kubelet summary API for Kubernetes metrics (based on [heapster](https://github.com/wavefronthq/wavefront-kubernetes-collector/tree/master/docs/heapster.md)). It additionally supports scraping Prometheus metrics format endpoints.

## Prerequisites
- Kubernetes 1.9+

## Configuration

The collector is plugin-driven and supports collecting metrics from multiple sources and writing metrics to Wavefront using a [Wavefront proxy](https://docs.wavefront.com/proxies.html) or via [direct ingestion](https://docs.wavefront.com/direct_ingestion.html).

See [configuration doc](https://github.com/wavefronthq/wavefront-kubernetes-collector/tree/master/docs/configuration.md) for detailed configuration information.

### Sources

Following sources are currently supported and can be configured using the `--source` flag:

1. Kubernetes source to collect performance metrics from the kubelet `/stats/summary` metrics API:
```
--source=kubernetes.summary_api:''
```
2. Prometheus source to scrape metrics from Prometheus metrics format endpoints such as kube state metrics:
```
--source=prometheus:''?url=http://kube-state-metrics.kube-system.svc.cluster.local:8080/metrics
```
Multiple prometheus sources can be added to scrape additional endpoints.

### Sending metrics to Wavefront

#### Using Wavefront Proxy

```
--sink=wavefront:?proxyAddress=wavefront-proxy.default.svc.cluster.local:2878&clusterName=k8s-cluster&includeLabels=true
```

#### Using Direct Ingestion
```
--sink=wavefront:?server=https://<YOUR_INSTANCE>.wavefront.com&token=<YOUR_TOKEN>&clusterName=k8s-cluster&includeLabels=true
```

## Installation

1. Clone this repo.
2. Edit the `wavefront` sink in `deploy/kubernetes/4-collector-deployment.yaml`.
3. Edit or remove the `prometheus` sink in the above file.
4. Run `kubectl apply -f deploy/kubernetes`

To verify the installation, find the pod name of the deployed `wavefront-collector` and run:

```
kubectl logs -f COLLECTOR_POD_NAME -n wavefront-collector
```
## Installation on OpenShift Origin 3.9

1. Login into Openshift master node.
2. Login into Openshift cluster `oc login -u <ADMIN_USER>`
3. Clone this Repo.
4. Create `wavefront-collector` namespace:
```
cd deploy/openshift
oc create -f collector-namespace.yaml
```
Note: Step 5 and 6 are needed if you are planning to use Wavefront Proxy else go to step 7
5. Login into Openshift web console and create storage under `wavefront-collector`. Select Access Mode as `RWX` and Size as `5 GiB`, give name to the storage and make a note of it.
6. Replace YOUR_CLUSTER, YOUR_API_TOKEN and STORAGE_NAME in wavefront-proxy.yaml and run:
```
oc create -f wavefront-proxy.yaml
```
7. Deploy kube-state-metrics into your cluster:
```
oc create -f kube-state.yaml
```
8. Edit the `wavefront` sink and cluster name in `4-collector-deployment.yaml` based on the metric sending mechanism as given below.
#### Using Wavefront Proxy

```
--sink=wavefront:?proxyAddress=wavefront-proxy.wavefront-collector.svc.cluster.local:2878&clusterName=openshift-cluster&includeLabels=true
```

#### Using Direct Ingestion
```
--sink=wavefront:?server=https://<YOUR_INSTANCE>.wavefront.com&token=<YOUR_TOKEN>&clusterName=openshift-cluster&includeLabels=true
```
9. Run below commands.
```
oc create -f 1-collector-cluster-role.yaml
oc create -f 2-collector-rbac.yaml
oc create -f 3-collector-service-account.yaml
oc create -f 4-collector-deployment.yaml
```

[ci-img]: https://travis-ci.com/wavefrontHQ/wavefront-kubernetes-collector.svg?branch=master
[ci]: https://travis-ci.com/wavefrontHQ/wavefront-kubernetes-collector
[go-report-img]: https://goreportcard.com/badge/github.com/wavefronthq/wavefront-kubernetes-collector
[go-report]: https://goreportcard.com/report/github.com/wavefronthq/wavefront-kubernetes-collector
[docker-pull-img]: https://img.shields.io/docker/pulls/wavefronthq/wavefront-kubernetes-collector.svg?logo=docker
[docker-img]: https://hub.docker.com/r/wavefronthq/wavefront-kubernetes-collector/
