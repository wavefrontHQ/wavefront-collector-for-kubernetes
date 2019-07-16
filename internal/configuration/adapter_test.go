package configuration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
)

func TestWavefrontSink(t *testing.T) {
	// proxy and direct ingestion
	conf := *buildWavefrontSink()
	uri, err := conf.convert()
	if err != nil {
		t.Errorf("error converting wavefront sink: %v", err)
	}
	assert.Equal(t, "wavefront", uri.Key)
	vals := uri.Val.Query()
	assert.True(t, len(vals) > 0)
	assert.Equal(t, 1, len(vals["proxyAddress"]))
	assert.Equal(t, 1, len(vals["server"]))
	assert.Equal(t, 1, len(vals["token"]))

	// tags, prefix and filters
	conf = WavefrontSinkConfig{
		Prefix: "k8s-metrics.",
		Tags:   map[string]string{"env": "test", "version": "1.14"},
		Filters: filter.Config{
			MetricWhitelist:    []string{"k8s.*", "kubernetes.*"},
			MetricBlacklist:    []string{"kops.*", "kube.*"},
			MetricTagWhitelist: map[string][]string{"env": {"prod"}},
			MetricTagBlacklist: map[string][]string{"env": {"dev", "test"}},
			TagInclude:         []string{"env"},
			TagExclude:         []string{"pod-template-id"},
		},
	}
	testCommon(t, conf, "wavefront")
}

func buildWavefrontSink() *WavefrontSinkConfig {
	return &WavefrontSinkConfig{
		ProxyAddress: "wf-proxy:2878",
		Server:       "https://foo.wavefront.com",
		Token:        "test",
	}
}

func TestKubeletSource(t *testing.T) {
	// default configuration (implicitly uses inClusterConfig)
	conf := SummaySourceConfig{
		Prefix: "kubernetes.",
	}
	uri, err := conf.convert()
	if err != nil {
		t.Errorf("error converting kubernetes source: %v", err)
	}
	assert.Equal(t, "kubernetes.summary_api", uri.Key)
	vals := uri.Val.Query()
	assert.Equal(t, 1, len(vals["prefix"]))
	assert.Equal(t, 0, len(vals["kubeletPort"]))
	assert.Equal(t, 0, len(vals["kubeletHttps"]))
	assert.Equal(t, 0, len(vals["inClusterConfig"]))
	assert.Equal(t, 0, len(vals["useServiceAccount"]))
	assert.Equal(t, 0, len(vals["insecure"]))
	assert.Equal(t, 0, len(vals["auth"]))

	// secure port configuration
	conf = *buildSummarySource()
	uri, err = conf.convert()
	if err != nil {
		t.Errorf("error converting kubernetes source: %v", err)
	}
	assert.Equal(t, "kubernetes.summary_api", uri.Key)
	assert.Equal(t, "https", uri.Val.Scheme)
	assert.Equal(t, "kubernetes.default.svc", uri.Val.Host)

	vals = uri.Val.Query()
	assert.Equal(t, 1, len(vals["useServiceAccount"]))
	assert.Equal(t, 1, len(vals["kubeletHttps"]))
	assert.Equal(t, 1, len(vals["kubeletPort"]))
	assert.Equal(t, 1, len(vals["insecure"]))

	assert.Equal(t, "true", vals["useServiceAccount"][0])
	assert.Equal(t, "true", vals["kubeletHttps"][0])
	assert.Equal(t, "10250", vals["kubeletPort"][0])
	assert.Equal(t, "true", vals["insecure"][0])

	// tags, prefix and filters
	conf = SummaySourceConfig{
		Prefix: "k8s-metrics.",
		Tags:   map[string]string{"env": "test", "version": "1.14"},
		Filters: filter.Config{
			MetricWhitelist:    []string{"k8s.*", "kubernetes.*"},
			MetricBlacklist:    []string{"kops.*", "kube.*"},
			MetricTagWhitelist: map[string][]string{"env": {"prod"}},
			MetricTagBlacklist: map[string][]string{"env": {"dev", "test"}},
			TagInclude:         []string{"env"},
			TagExclude:         []string{"pod-template-id"},
		},
	}
	testCommon(t, conf, "kubernetes.summary_api")
}

func buildSummarySource() *SummaySourceConfig {
	return &SummaySourceConfig{
		URL:               "https://kubernetes.default.svc",
		UseServiceAccount: "true",
		KubeletHttps:      "true",
		KubeletPort:       "10250",
		Insecure:          "true",
	}
}

func TestPrometheusSource(t *testing.T) {
	conf := *buildPromSource()
	uri, err := conf.convert()
	if err != nil {
		t.Errorf("error converting prometheus source: %v", err)
	}
	assert.Equal(t, "prometheus", uri.Key)
	vals := uri.Val.Query()
	assert.Equal(t, 1, len(vals["url"]))
	assert.Equal(t, 1, len(vals["source"]))
	assert.Equal(t, "http://1.2.3.4:8080/metrics", vals["url"][0])
	assert.Equal(t, "test_source", vals["source"][0])

	conf = PrometheusSourceConfig{
		Prefix: "k8s-metrics.",
		Tags:   map[string]string{"env": "test", "version": "1.14"},
		Filters: filter.Config{
			MetricWhitelist:    []string{"k8s.*", "kubernetes.*"},
			MetricBlacklist:    []string{"kops.*", "kube.*"},
			MetricTagWhitelist: map[string][]string{"env": {"prod"}},
			MetricTagBlacklist: map[string][]string{"env": {"dev", "test"}},
			TagInclude:         []string{"env"},
			TagExclude:         []string{"pod-template-id"},
		},
	}
	testCommon(t, conf, "prometheus")
}

func buildPromSource() *PrometheusSourceConfig {
	return &PrometheusSourceConfig{
		URL:    "http://1.2.3.4:8080/metrics",
		Source: "test_source",
	}
}

func TestTelegrafSource(t *testing.T) {
	conf := TelegrafSourceConfig{
		Plugins:    []string{"mem", "disk", "diskio"},
		Collection: CollectionSourceConfig{Interval: "1m", TimeOut: "2m"},
	}
	uri, err := conf.convert()
	if err != nil {
		t.Errorf("error converting telegraf source: %v", err)
	}
	assert.Equal(t, "telegraf", uri.Key)
	vals := uri.Val.Query()
	assert.Equal(t, 1, len(vals["plugins"]))
	assert.Equal(t, "1m", vals["collectionInterval"][0])
	assert.Equal(t, "2m", vals["timeOut"][0])

	conf = TelegrafSourceConfig{
		Plugins: []string{"mem", "disk", "diskio"},
	}
	uri, err = conf.convert()
	if err != nil {
		t.Errorf("error converting telegraf source: %v", err)
	}
	assert.Equal(t, "telegraf", uri.Key)
	vals = uri.Val.Query()
	assert.Equal(t, 1, len(vals["plugins"]))
	assert.Equal(t, 0, len(vals["collectionInterval"]))
	assert.Equal(t, 0, len(vals["timeOut"]))
}

func buildTelegrafSource() *TelegrafSourceConfig {
	return &TelegrafSourceConfig{
		Plugins: []string{"mem", "disk", "diskio"},
	}
}

func TestSystemdSource(t *testing.T) {
	conf := *buildSystemdSource()
	uri, err := conf.convert()
	if err != nil {
		t.Errorf("error converting systemd source: %v", err)
	}
	assert.Equal(t, "systemd", uri.Key)
	vals := uri.Val.Query()
	assert.Equal(t, 1, len(vals["taskMetrics"]))
	assert.Equal(t, 1, len(vals["restartMetrics"]))
	assert.Equal(t, 1, len(vals["startTimeMetrics"]))

	conf = SystemdSourceConfig{
		Prefix: "k8s-metrics.",
		Tags:   map[string]string{"env": "test", "version": "1.14"},
		Filters: filter.Config{
			MetricWhitelist:    []string{"k8s.*", "kubernetes.*"},
			MetricBlacklist:    []string{"kops.*", "kube.*"},
			MetricTagWhitelist: map[string][]string{"env": {"prod"}},
			MetricTagBlacklist: map[string][]string{"env": {"dev", "test"}},
			TagInclude:         []string{"env"},
			TagExclude:         []string{"pod-template-id"},
		},
	}
	testCommon(t, conf, "systemd")
}

func buildSystemdSource() *SystemdSourceConfig {
	return &SystemdSourceConfig{
		IncludeTaskMetrics:      true,
		IncludeRestartMetrics:   true,
		IncludeStartTimeMetrics: true,
	}
}

func TestFullConversion(t *testing.T) {
	conf := Config{
		PushInterval:          2 * time.Minute,
		SinkExportDataTimeout: 20 * time.Second,
		MaxProcs:              4,
		EnableDiscovery:       true,
		ClusterName:           "prod-cluster",

		Sinks:             []*WavefrontSinkConfig{buildWavefrontSink()},
		SummaryConfig:     buildSummarySource(),
		PrometheusConfigs: []*PrometheusSourceConfig{buildPromSource()},
		TelegrafConfigs:   []*TelegrafSourceConfig{buildTelegrafSource()},
		SystemdConfig:     buildSystemdSource(),
	}
	opts, err := conf.Convert()
	if err != nil {
		t.Errorf("full conversion error: %v", err)
	}
	assert.Equal(t, 2*time.Minute, opts.PushInterval)
	assert.Equal(t, 20*time.Second, opts.SinkExportDataTimeout)
	assert.Equal(t, 4, opts.MaxProcs)
	assert.True(t, opts.EnableDiscovery)
	assert.True(t, len(opts.Sinks) == 1)
	assert.True(t, len(opts.Sources) == 4)

	// confirm cluster is carried over from top level to sink
	sinkVals := opts.Sinks[0].Val.Query()
	assert.Equal(t, "prod-cluster", sinkVals["clusterName"][0])
}

func testCommon(t *testing.T, a adapter, key string) {
	// Tags, Prefix and Filtering properties common to all configuration types
	uri, err := a.convert()
	if err != nil {
		t.Errorf("error converting %s: %v", key, err)
	}

	vals := uri.Val.Query()
	assert.Equal(t, key, uri.Key)

	assert.True(t, len(vals["tag"]) > 0)
	assert.Equal(t, 1, len(vals["prefix"]))
	assert.Equal(t, 2, len(vals[filter.MetricWhitelist]))
	assert.Equal(t, 2, len(vals[filter.MetricBlacklist]))
	assert.Equal(t, 1, len(vals[filter.MetricTagWhitelist]))
	assert.Equal(t, 1, len(vals[filter.MetricTagBlacklist]))
	assert.Equal(t, 1, len(vals[filter.TagInclude]))
	assert.Equal(t, 1, len(vals[filter.TagExclude]))
}
