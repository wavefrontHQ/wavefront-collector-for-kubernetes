# wavefront-collector-kubernetes [![build status][ci-img]][ci] [![Go Report Card][go-report-img]][go-report] [![Docker Pulls][docker-pull-img]][docker-img]

[Wavefront](https://docs.wavefront.com) is a high-performance streaming analytics platform for monitoring and optimizing your environment and applications.

The Wavefront Collector for Kubernetes enables monitoring Kubernetes clusters and sending metrics to Wavefront.

## Features
* Collects real-time metrics from all layers of a Kubernetes environment
* Multiple sources of metrics providing comprehensive insight:
  - Kubernetes source: For [core kubernetes metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#kubernetes-source)
  - Prometheus source: For scraping prometheus metric endpoints (API server, etcd, NGINX etc)
  - Telegraf source: For [host level metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#telegraf-source)
  - Systemd source: For [host level systemd metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#systemd-source)
* [Auto discovery](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/discovery.md) of pods and services based on annotation and configuration
* Daemonset mode for high scalability with leader election for monitoring cluster level resources
* Rich [filtering](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/filtering.md) support
* Auto reload of configuration changes
* [Internal metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#collector-health-metrics) for tracking the collector health and configuration

## Prerequisites
- Kubernetes 1.9+

## Installation

Refer to the [installation instructions](https://docs.wavefront.com/kubernetes.html#kubernetes-setup).

## Configuration

The installation instructions use a default configuration suitable for most use cases. The collector supports various advanced configuration options. Configuration options have changed in 1.0 (upcoming release). See the documentation for the version of the collector you are running:

| Version | Documentation |
| ----- | -------- |
| `1.x` | [Docs](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/tree/master/docs) |
| `0.9.x` | [Docs](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/tree/v0.9.8/docs) |

## OpenShift
This collector supports monitoring of Openshift Origin 3.9 clusters. See [openshift.md](https://github.com/wavefronthq/wavefront-kubernetes-collector/tree/master/docs/openshift.md) for detailed installation instructions.

## Contributing
Public contributions are always welcome. Please feel free to report issues or submit pull requests.

[ci-img]: https://travis-ci.com/wavefrontHQ/wavefront-kubernetes-collector.svg?branch=master
[ci]: https://travis-ci.com/wavefrontHQ/wavefront-kubernetes-collector
[go-report-img]: https://goreportcard.com/badge/github.com/wavefronthq/wavefront-kubernetes-collector
[go-report]: https://goreportcard.com/report/github.com/wavefronthq/wavefront-kubernetes-collector
[docker-pull-img]: https://img.shields.io/docker/pulls/wavefronthq/wavefront-kubernetes-collector.svg?logo=docker
[docker-img]: https://hub.docker.com/r/wavefronthq/wavefront-kubernetes-collector/
