module github.com/wavefronthq/wavefront-collector-for-kubernetes

go 1.15

require (
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/armon/go-metrics v0.3.0 // indirect
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/couchbase/go-couchbase v0.0.0-20191115001346-d9e5b3bd1ebc // indirect
	github.com/couchbase/gomemcached v0.0.0-20191004160342-7b5da2ec40b2 // indirect
	github.com/couchbase/goutils v0.0.0-20191018232750-b49639060d85 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-redis/redis v6.15.6+incompatible // indirect
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/cadvisor v0.35.0
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/hashicorp/consul/api v1.3.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-rootcerts v1.0.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/memberlist v0.1.5 // indirect
	github.com/hashicorp/serf v0.8.5 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/influxdata/telegraf v0.10.2-0.20191023195903-9a4f08e94774
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgx v3.6.0+incompatible // indirect
	github.com/json-iterator/go v1.1.12
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lib/pq v1.3.0 // indirect
	github.com/miekg/dns v1.1.26 // indirect
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.9.1
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/shirou/gopsutil v2.20.6+incompatible
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.7.4 // indirect
	github.com/wavefronthq/go-metrics-wavefront v1.0.3
	github.com/wavefronthq/wavefront-sdk-go v0.9.9
	golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.17.1
	k8s.io/apiserver v0.15.7
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200410163147-594e756bea31 // indirect
	k8s.io/kubernetes v1.18.15
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
