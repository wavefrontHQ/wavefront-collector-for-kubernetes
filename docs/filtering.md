# Filtering Metrics

The Wavefront Kubernetes Collector supports filtering metrics before they are reported to Wavefront. The following filtering options are supported:

  * **metricWhitelist**: List of glob patterns. Only metrics with names matching the whitelist are reported.
  * **metricBlacklist** List of glob patterns. Metrics with names matching the blacklist are dropped.
  * **metricTagWhitelist** Map of tag names to list of glob patterns. Only metrics containing tag keys and values matching the whitelist will be reported.
  * **metricTagBlacklist** Map of tag names to list of glob patterns. Metrics containing blacklisted tag keys and values will be dropped.
  * **tagInclude** List of glob patterns. Tags with matching keys will be included. All other tags will be excluded.
  * **tagExclude** List of glob patterns. Tags with matching keys will be excluded.

Filtering can be enabled on the sink or prometheus sources:
```
--sink=wavefront:?proxyAddress=wavefront-proxy.default.svc.cluster.local:2878&clusterName=k8s-cluster&includeLabels=true&metricBlacklist=kube.dns.go*&metricBlacklist=kube.apiserver.*&metricTagWhitelist=env:[prod1*,prod2*]
```

```
--source=prometheus:''?url=http://svcname.svc.cluster.local:8080/metrics&metricWhitelist=a.b.*
```

Filtering is also supported within discovery configurations:
```
global:
  discovery_interval: 5m
prom_configs:
  - name: kube-dns-discovery
    labels:
      k8s-app: kube-dns
    port: 10054
    prefix: kube.dns.
    tags:
      type: test
    filters:
      metricBlacklist:
      - 'kube.dns.go.*'
      - 'kube.dns.probe.*'
      metricTagWhitelist:
        env:
        - 'prod1*'
        - 'prod2*'
        service:
        - 'app1*'
        - '?app2*'
```
