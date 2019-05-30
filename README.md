# wavefront-kubernetes-collector [![build status][ci-img]][ci] [![Go Report Card][go-report-img]][go-report] [![Docker Pulls][docker-pull-img]][docker-img]

[Wavefront](https://docs.wavefront.com) is a high-performance streaming analytics platform for monitoring and optimizing your environment and applications.

The Wavefront Kubernetes Collector enables monitoring Kubernetes clusters and sending metrics to Wavefront.

## Features
* Collects real-time metrics from all layers of a Kubernetes environment
* Multiple sources of metrics providing comprehensive insight
  - Kubernetes source: For [core kubernetes metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#kubernetes-source)
  - Prometheus source: For scraping prometheus metric endpoints (API server, etcd, NGINX etc)
  - Telegraf source: For [host level metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#telegraf-source)
  - Systemd source: For [host level systemd metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#systemd-source)
* Annotation and configuration based [auto discovery](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/discovery.md) of pods and services
* Daemonset mode for high scalability
* Rich [filtering](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/filtering.md) support.
* Auto reload of configuration changes
* Emits internal [health metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#collector-health-metrics) for tracking the state of your collector deployments

## Prerequisites
- Kubernetes 1.9+

## Configuration

The collector is plugin-driven and supports collecting metrics from multiple sources and writing metrics to Wavefront using a [Wavefront proxy](https://docs.wavefront.com/proxies.html) or via [direct ingestion](https://docs.wavefront.com/direct_ingestion.html). See the [configuration doc](https://github.com/wavefronthq/wavefront-kubernetes-collector/tree/master/docs/configuration.md) for detailed configuration information.

### Sources

Multiple sources are supported and can be configured using the `--source` flag.

For example, to configure the Kubernetes source:
```
--source=kubernetes.summary_api:''
```

To configure a Prometheus source to scrape metrics from a kube-state-metrics endpoint:
```
--source=prometheus:''?url=http://kube-state-metrics.kube-system.svc.cluster.local:8080/metrics
```
Multiple prometheus sources can be added to scrape additional endpoints.

### Auto Discovery
The collector can auto discover pods and services that export Prometheus format metrics. See the [discovery documentation](https://github.com/wavefronthq/wavefront-kubernetes-collector/tree/master/docs/discovery.md) for details.

### Sending metrics to Wavefront

#### Using Wavefront Proxy

```
--sink=wavefront:?proxyAddress=wavefront-proxy.default.svc.cluster.local:2878&clusterName=k8s-cluster&includeLabels=true
```

#### Using Direct Ingestion
```
--sink=wavefront:?server=https://<YOUR_INSTANCE>.wavefront.com&token=<YOUR_TOKEN>&clusterName=k8s-cluster&includeLabels=true
```

## Deployment

The collector can be deployed as a daemonset (per node agent) or as a single cluster level agent.

1. Clone this repo.
2. Retain the `deploy/kubernetes/4-collector-daemonset.yaml` or `deploy/kubernetes/4-collector-deployment.yaml` depending on which mode you'd like to deploy the collector as. Remove the other file.
4. Run `kubectl apply -f deploy/kubernetes`

To verify the installation, run:

```
kubectl get pods -n wavefront-collector
```

## OpenShift
This collector supports monitoring of Openshift Origin 3.9 clusters. See [openshift.md](https://github.com/wavefronthq/wavefront-kubernetes-collector/tree/master/docs/openshift.md) for detailed installation instructions.

[ci-img]: https://travis-ci.com/wavefrontHQ/wavefront-kubernetes-collector.svg?branch=master
[ci]: https://travis-ci.com/wavefrontHQ/wavefront-kubernetes-collector
[go-report-img]: https://goreportcard.com/badge/github.com/wavefronthq/wavefront-kubernetes-collector
[go-report]: https://goreportcard.com/report/github.com/wavefronthq/wavefront-kubernetes-collector
[docker-pull-img]: https://img.shields.io/docker/pulls/wavefronthq/wavefront-kubernetes-collector.svg?logo=docker
[docker-img]: https://hub.docker.com/r/wavefronthq/wavefront-kubernetes-collector/
