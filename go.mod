module github.com/wavefronthq/wavefront-collector-for-kubernetes

go 1.15

require (
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/coreos/go-systemd/v22 v22.1.0
	github.com/couchbase/go-couchbase v0.0.0-20191115001346-d9e5b3bd1ebc // indirect
	github.com/couchbase/gomemcached v0.0.0-20191004160342-7b5da2ec40b2 // indirect
	github.com/couchbase/goutils v0.0.0-20191018232750-b49639060d85 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-redis/redis v6.15.6+incompatible // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/cadvisor v0.34.0
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/hashicorp/consul v1.4.5 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-rootcerts v1.0.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/serf v0.8.5 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/influxdata/telegraf v1.14.0
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/json-iterator/go v1.1.9
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.9.1
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563
	github.com/shirou/gopsutil v2.20.2+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/gjson v1.3.4 // indirect
	github.com/wavefronthq/go-metrics-wavefront v1.0.2
	github.com/wavefronthq/wavefront-sdk-go v0.9.5
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/yaml.v2 v2.2.5
	k8s.io/api v0.15.7
	k8s.io/apimachinery v0.17.1
	k8s.io/apiserver v0.15.7
	k8s.io/client-go v0.15.7
	k8s.io/kubernetes v1.16.3
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
)

exclude (
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/cloud-provider v0.0.0
	k8s.io/cluster-bootstrap v0.0.0
	k8s.io/code-generator v0.0.0
	k8s.io/component-base v0.0.0
	k8s.io/cri-api v0.0.0
	k8s.io/csi-translation-lib v0.0.0
	k8s.io/kube-aggregator v0.0.0
	k8s.io/kube-controller-manager v0.0.0
	k8s.io/kube-proxy v0.0.0
	k8s.io/kube-scheduler v0.0.0
	k8s.io/kubectl v0.0.0
	k8s.io/kubelet v0.0.0
	k8s.io/legacy-cloud-providers v0.0.0
	k8s.io/metrics v0.0.0
	k8s.io/sample-apiserver v0.0.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.1
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190116210010-30d6a91f580b
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
)
