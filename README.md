# Wavefront Collector for Kubernetes
[![Go Report Card][go-report-img]][go-report] [![Docker Pulls][docker-pull-img]][docker-img]

[Wavefront](https://docs.wavefront.com) is a high-performance streaming analytics platform for monitoring and optimizing your environment and applications.

The Wavefront Collector for Kubernetes is an agent that runs as a [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) on each node within a Kubernetes cluster. It collects metrics and events about the cluster and sends them to the Wavefront SaaS service.

## Features
* Collects real-time data from all layers of a Kubernetes environment
* Multiple sources of metrics providing comprehensive insight:
  - Kubernetes (kubelet) source: For [core kubernetes metrics](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/blob/main/docs/metrics.md#kubernetes-source)
  - Prometheus source: For scraping prometheus metric endpoints (API server, etcd, NGINX etc)
  - Kubernetes state source: For [resource state metrics](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/blob/main/docs/metrics.md#kubernetes-state-source)    
  - Telegraf source: For [host and application](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/blob/main/docs/metrics.md#telegraf-source) level metrics
  - Systemd source: For [host level systemd metrics](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/blob/main/docs/metrics.md#systemd-source)
* [Auto discovery](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/blob/main/docs/discovery.md) of pods and services based on annotation and configuration
* Daemonset mode for high scalability with leader election for monitoring cluster level resources
* Rich [filtering](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/blob/main/docs/filtering.md) support
* Auto reload of configuration changes
* [Internal metrics](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/blob/main/docs/metrics.md#collector-health-metrics) for tracking the collector health and configuration

## Installation

Refer to the [installation instructions](https://docs.wavefront.com/kubernetes.html#kubernetes-quick-install-using-the-kubernetes-operator).

## Configuration

The installation instructions use a default configuration suitable for most use cases. Refer to the [documentation](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/tree/main/docs) for details on all the configuration options.

## Building

Build using `make` and the provided `Makefile`. 

Commonly used `make` options include: 
* `fmt` to `go fmt` all your code
* `tests` to run all the unit tests 
* `build` that creates a local executable
* `container` that uses a docker container to build for consistency and reproducability 

## Troubleshooting Dropped Metrics

Formerly, we would see the following error in the proxy logs when a metric has too many tags: `Too many point tags`.
However, logic has been added to the collector to automatically drop tags in priority order
to ensure that metrics make it through to the proxy and no longer cause this error.
This is the order of the logic used to drop tags:
1. Explicitly excluded tags (from `tagExclude` config).
   Refer [here](https://github.com/wavefrontHQ/wavefront-operator-for-kubernetes/blob/main/deploy/kubernetes/scenarios/wavefront-full-config.yaml) for an example scenario.
1. Tags are empty or are interpreted to be empty (`"tag.key": ""`, `"tag.key": "-"`, or `"tag.key": "/"`).
1. Tags are explicitly excluded
   (`"namespace_id": "..."`, `"host_id": "..."`, `"pod_id": "..."`, or `"hostname": "..."`).
1. Tag **values** are duplicated, and the shorter key is kept
   (`"tag.key": "same value"` is kept instead of `"tag.super.long.key": "same value"`).
1. Tag key matches `alpha.*` or `beta.*`, after keys have been sorted
   (e.g. `"alpha.eksctl.io/nodegroup-name": "arm-group"` or `"beta.kubernetes.io/arch": "amd64"`).
1. (As a last resort) tag key matches IaaS-specific tags, after keys have been sorted
   (`"kubernetes.azure.com/agentpool": "agentpool"`, `"topology.gke.io/zone": "us-central1-c"`, `"eksctl.io/nodegroup-name": "arm-group"`, etc.).

## Contributing
Public contributions are always welcome. Please feel free to report issues or submit pull requests.

[go-report-img]: https://goreportcard.com/badge/github.com/wavefronthq/wavefront-kubernetes-collector
[go-report]: https://goreportcard.com/report/github.com/wavefronthq/wavefront-kubernetes-collector
[docker-pull-img]: https://img.shields.io/docker/pulls/wavefronthq/wavefront-kubernetes-collector.svg?logo=docker
[docker-img]: https://hub.docker.com/r/wavefronthq/wavefront-kubernetes-collector/
