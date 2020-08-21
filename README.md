# Wavefront Collector for Kubernetes
[![build status][ci-img]][ci] [![Go Report Card][go-report-img]][go-report] [![Docker Pulls][docker-pull-img]][docker-img]

[Wavefront](https://docs.wavefront.com) is a high-performance streaming analytics platform for monitoring and optimizing your environment and applications.

The Wavefront Collector for Kubernetes is an agent that runs as a [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) on each node within a Kubernetes cluster. It collects metrics and events about the cluster and sends them to the Wavefront SaaS service.

## Features
* Collects real-time data from all layers of a Kubernetes environment
* Multiple sources of metrics providing comprehensive insight:
  - Kubernetes (kubelet) source: For [core kubernetes metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#kubernetes-source)
  - Prometheus source: For scraping prometheus metric endpoints (API server, etcd, NGINX etc)
  - Kubernetes state source: For [resource state metrics](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/blob/master/docs/metrics.md#kubernetes-state-source)    
  - Telegraf source: For [host and application](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#telegraf-source) level metrics
  - Systemd source: For [host level systemd metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#systemd-source)
* [Auto discovery](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/discovery.md) of pods and services based on annotation and configuration
* Daemonset mode for high scalability with leader election for monitoring cluster level resources
* Rich [filtering](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/filtering.md) support
* Auto reload of configuration changes
* [Internal metrics](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/metrics.md#collector-health-metrics) for tracking the collector health and configuration

## Installation

Refer to the [installation instructions](https://docs.wavefront.com/kubernetes.html#kubernetes-setup).

## Configuration

The installation instructions use a default configuration suitable for most use cases. Refer to the [documentation](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/tree/master/docs) for details on all the configuration options.

## Contributing
Public contributions are always welcome. Please feel free to report issues or submit pull requests.

[ci-img]: https://travis-ci.com/wavefrontHQ/wavefront-collector-for-kubernetes.svg?branch=master
[ci]: https://travis-ci.com/wavefrontHQ/wavefront-collector-for-kubernetes
[go-report-img]: https://goreportcard.com/badge/github.com/wavefronthq/wavefront-kubernetes-collector
[go-report]: https://goreportcard.com/report/github.com/wavefronthq/wavefront-kubernetes-collector
[docker-pull-img]: https://img.shields.io/docker/pulls/wavefronthq/wavefront-kubernetes-collector.svg?logo=docker
[docker-img]: https://hub.docker.com/r/wavefronthq/wavefront-kubernetes-collector/
