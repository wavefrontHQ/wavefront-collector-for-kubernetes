# Filtering

## Table of Contents
* [Introduction](#introduction)
* [Metrics Filtering](#metrics-filtering)
* [Events Filtering](#events-filtering)

## Introduction
The Wavefront Collector supports filtering metrics and events. Filters are based on [glob patterns](https://github.com/gobwas/glob#syntax) (similar to standard wildcards).

## Metrics Filtering

The Wavefront Collector for Kubernetes supports filtering metrics before they are reported to Wavefront. The following filtering options are supported for metrics:

  * **metricAllowList**: List of glob patterns. Only metrics with names matching this list are reported.
  * **metricDenyList**: List of glob patterns. Metrics with names matching this list are dropped.
  * **metricTagAllowList**: Map of tag names to list of glob patterns. Only metrics containing tag keys and values matching this list will be reported.
  * **metricTagDenyList**: Map of tag names to list of glob patterns. Metrics containing these tag keys and values will be dropped.
  * **tagInclude**: List of glob patterns. Tags with matching keys will be included. All other tags will be excluded.
  * **tagExclude**: List of glob patterns. Tags with matching keys will be excluded.

Filtering can be enabled on the sink or sources. Where it is applied controls the scope of metrics towards which the filtering applies.

Filtering applied on the sink applies to all the metrics collected by the collector:

```yaml
sinks:
  - proxyAddress: wavefront-proxy.default.svc.cluster.local:2878

  # global sink level filter
  filters:
    # Filter out all go runtime metrics for kube-dns and apiserver.
    metricDenyList:
    - 'kube.dns.go.*'
    - 'kube.apiserver.go.*'

    # Allow metrics that have an environment tag of production or staging
    metricTagAllowList:
      env:
      - 'prod*'
      - 'staging*'

    # Block metrics that have an environment tag of test.
    metricTagDenyList:
      env:
      - 'test*'
```

Filtering applied on a source applies only to metrics collected by that source:
```yaml
prometheus_sources:
  # collect metrics from a prometheus endpoint
  - url: 'http://prom-endpoint.default.svc.cluster.local:9153/metrics'
    prefix: 'prom.app.'

    filters:
      # Filter out all go runtime metrics
      metricDenyList:
      - 'prom.app.go.*'

      # Allow metrics that have an environment tag of production or staging
      metricTagAllowList:
        env:
        - 'prod*'
        - 'staging*'

      # Block metrics that have an environment tag of test.
      metricTagDenyList:
        env:
        - 'test*'
```

Filtering can also be specified within discovery rules, and only apply towards the metrics collected from the discovered targets:
```yaml
discovery_configs:
  - name: kube-dns-discovery
    labels:
      k8s-app: kube-dns
    port: 10054
    prefix: kube.dns.

    # filtering rules to be applied towards kube-dns metrics
    filters:
      metricDenyList:
      - 'kube.dns.go.*'
      - 'kube.dns.probe.*'

      metricTagAllowList:
        env:
        - 'prod1*'
        - 'prod2*'
        service:
        - 'app1*'
        - '?app2*'
```

## Events Filtering

The Wavefront Collector for Kubernetes also supports filtering events before they are reported to Wavefront. The following filtering options are supported for events:

* **tagAllowList**: Map of tag names to list of glob patterns. Only events containing tag keys and values matching the list will be reported.
* **tagDenyList**: Map of tag names to list of glob patterns. Events containing these tag keys and values will be dropped.
* **tagAllowListSets**: List of maps of tag names to list of glob patterns. Filters within each map are AND'd. Filters between the maps are OR'd.
* **tagDenyListSets**: List of maps of tag names to list of glob patterns. Filters within each map are AND'd. Filters between the maps are OR'd.

Event filtering is specified within the top level events section in the config. For example to allow either Pod or DaemonSet events with a specific reason:

```yaml
events:
  filters:
    tagAllowListSets:
    - kind:
     - "Pod"
     reason:
     - "Scheduled"
     - "Failed*"
    - kind:
     - "DaemonSet"
     reason:
     - "SuccessfulCreate"
     - "Failed*"
```
