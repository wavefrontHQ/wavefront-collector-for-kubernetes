# wavefront-kubernetes-collector

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

3. OpenShift source to collect performance metrics from the kubelet `/stats/summary` metrics API:
```
--source=kubernetes.summary_api:https://172.0.0.1:10250
```
4. OKE does not suppor performance metrics from the kubelet `/stats/summary` metrics API:
```
--source=kubernetes.summary_api:''
```


### Sending metrics to Wavefront

#### Using Wavefront Proxy

```
--sink=wavefront:?proxyAddress=wavefront-proxy.default.svc.cluster.local:2878&clusterName=k8s-cluster&includeLabels=true
```

#### Using Direct Ingestion
```
--sink=wavefront:?server=https://<YOUR_INSTANCE>.wavefront.com&token=<YOUR_TOKEN>&clusterName=k8s-cluster&includeLabels=true
```

## Installation on Kubernetes

1. Clone this repo.
2. Edit the `wavefront` sink in `deploy/kubernetes/4-collector-deployment.yaml`.
3. Edit or remove the `prometheus` sink in the above file.
4. Run `kubectl apply -f deploy/kubernetes`

To verify the installation, find the pod name of the deployed `wavefront-collector` and run:

```
kubectl logs -f COLLECTOR_POD_NAME -n wavefront-collector
```

## Installation on OpenShift

1. On OpenShift web console create a new project called `wavefront-collector`
1. Deploy [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) on your Openshift
1. Clone this repo.
1. Edit the `wavefront` sink in `deploy/openshift/4-collector-deployment.yaml`.
1. Edit the `kubernetes.summary_api` sink in the above file if you are running OKD.
1. Run `kubectl apply -f deploy/openshift`

To verify the installation, find the running pod on the web console and take a look of the logs.
