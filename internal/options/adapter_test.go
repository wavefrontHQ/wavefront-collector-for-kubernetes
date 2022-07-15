// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"fmt"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/flags"

	"github.com/stretchr/testify/assert"
)

func TestGetTransforms(t *testing.T) {
	vals := make(map[string][]string)
	vals["prefix"] = []string{"test."}
	vals["source"] = []string{"source."}
	vals["tag"] = []string{"key1:val1", "key2:val2"}
	vals["metricWhitelist"] = []string{"kube.*", "kubernetes.*"}
	vals["metricBlacklist"] = []string{"kube.go.*", "kube.http.*"}

	tr := getTransforms(vals)
	assert.Equal(t, "test.", tr.Prefix)
	assert.Equal(t, "source.", tr.Source)
	assert.Equal(t, 2, len(tr.Tags))
	assert.Equal(t, "val1", tr.Tags["key1"])
	assert.Equal(t, "val2", tr.Tags["key2"])

	assert.Equal(t, 2, len(tr.Filters.MetricWhitelist))
	assert.Equal(t, 2, len(tr.Filters.MetricBlacklist))
}

func TestAddSummarySource(t *testing.T) {
	uri, err := buildSummarySource()
	assert.NoError(t, err)

	cfg := &configuration.Config{Sources: &configuration.SourceConfig{}}
	addSummarySource(cfg, uri)

	summ := cfg.Sources.SummaryConfig
	assert.Equal(t, "https://kubernetes.default.svc", summ.URL)
	assert.Equal(t, "10250", summ.KubeletPort)
	assert.Equal(t, "true", summ.KubeletHttps)
	assert.Equal(t, "true", summ.InClusterConfig)
	assert.Equal(t, "true", summ.UseServiceAccount)
	assert.Equal(t, "true", summ.Insecure)
	assert.Equal(t, "kubernetes.", summ.Prefix)
	assert.Equal(t, 2, len(summ.Tags))
}

func TestAddStateSource(t *testing.T) {
	uri, err := buildStateSource("kstate.")
	assert.NoError(t, err)

	cfg := &configuration.Config{Sources: &configuration.SourceConfig{}}
	addStateSource(cfg, uri)

	state := cfg.Sources.StateConfig
	assert.Equal(t, "kstate.", state.Prefix)
	assert.Equal(t, 2, len(state.Tags))

	// without any prefix
	uri, err = buildStateSource("")
	assert.NoError(t, err)

	cfg = &configuration.Config{Sources: &configuration.SourceConfig{}}
	addStateSource(cfg, uri)

	state = cfg.Sources.StateConfig
	assert.Equal(t, "", state.Prefix)
	assert.Equal(t, 2, len(state.Tags))
}

func buildStateSource(prefix string) (flags.Uri, error) {
	values := url.Values{}
	if prefix != "" {
		addVal(values, "prefix", prefix)
	}
	encodeTags(values, "", map[string]string{"k1": "v1", "k2": "v2"})
	return buildUri("kubernetes.state", "", values.Encode())
}

func TestAddPrometheusSource(t *testing.T) {
	values := url.Values{}
	values["url"] = []string{"http://kube-state-metrics.kube-system.svc.cluster.local:8080/metrics"}
	values["prefix"] = []string{"prom."}
	encodeTags(values, "", map[string]string{"k1": "v1", "k2": "v2"})

	uri, err := buildUri("prometheus", "", values.Encode())
	assert.NoError(t, err)

	cfg := &configuration.Config{Sources: &configuration.SourceConfig{}}
	addPrometheusSource(cfg, uri)

	assert.True(t, len(cfg.Sources.PrometheusConfigs) == 1)

	pcfg := cfg.Sources.PrometheusConfigs[0]

	assert.Equal(t, "http://kube-state-metrics.kube-system.svc.cluster.local:8080/metrics", pcfg.URL)
	assert.Equal(t, "prom.", pcfg.Prefix)
	assert.Equal(t, 2, len(pcfg.Tags))
}

func TestAddCadvisorSource(t *testing.T) {
	values := url.Values{}
	values["prefix"] = []string{"kubernetes.cadvisor."}

	uri, err := buildUri("kubernetes.cadvisor", "", values.Encode())
	assert.NoError(t, err)

	cfg := &configuration.Config{Sources: &configuration.SourceConfig{}}
	addCadvisorSource(cfg, uri)

	assert.NotNil(t, cfg.Sources.CadvisorConfig)

	cadvisorCfg := cfg.Sources.CadvisorConfig

	assert.Equal(t, "kubernetes.cadvisor.", cadvisorCfg.Prefix)
}

func buildSummarySource() (flags.Uri, error) {
	values := url.Values{}
	addVal(values, "kubeletPort", "10250")
	addVal(values, "kubeletHttps", "true")
	addVal(values, "inClusterConfig", "true")
	addVal(values, "useServiceAccount", "true")
	addVal(values, "insecure", "true")
	addVal(values, "auth", "")
	addVal(values, "prefix", "kubernetes.")
	encodeTags(values, "", map[string]string{"k1": "v1", "k2": "v2"})

	kurl := "https://kubernetes.default.svc"
	return buildUri("kubernetes.summary_api", kurl, values.Encode())
}

func buildWavefrontSink() (flags.Uri, error) {
	values := url.Values{}
	addVal(values, "proxyAddress", "wavefront-proxy.default.svc.cluster.local:2878")
	addVal(values, "clusterName", "test-cluster")
	addVal(values, "includeLabels", "true")
	addVal(values, "prefix", "staging.")
	return buildUri("wavefront", "", values.Encode())
}

func TestConvert(t *testing.T) {
	opts := &CollectorRunOptions{
		AgentType:             AllAgentType,
		MetricResolution:      120 * time.Second,
		SinkExportDataTimeout: 130 * time.Second,
		EnableDiscovery:       true,
	}
	uri, err := buildSummarySource()
	assert.NoError(t, err)
	opts.Sources = append(opts.Sources, uri)
	uri, err = buildWavefrontSink()
	assert.NoError(t, err)
	opts.Sinks = append(opts.Sinks, uri)

	cfg, err := opts.Convert()
	assert.NoError(t, err)

	assert.True(t, cfg.ScrapeCluster)
	assert.True(t, cfg.EnableDiscovery)
	assert.Equal(t, 120*time.Second, cfg.DefaultCollectionInterval)
	assert.Equal(t, 120*time.Second, cfg.FlushInterval)
	assert.Equal(t, 130*time.Second, cfg.SinkExportDataTimeout)
	assert.Equal(t, "test-cluster", cfg.ClusterName)

	assert.Equal(t, 1, len(cfg.Sinks))
	assert.NotNil(t, cfg.Sources.SummaryConfig)
	assert.NotNil(t, cfg.Sources.StatsConfig)
	assert.Empty(t, cfg.Sinks[0].Prefix)
	assert.Equal(t, "staging.", cfg.Sources.SummaryConfig.Prefix)
}

func buildUri(key, address, rawQuery string) (flags.Uri, error) {
	u, err := url.Parse(address + "?")
	if err != nil {
		return flags.Uri{}, err
	}
	u.RawQuery = rawQuery

	uri := flags.Uri{Key: key, Val: *u}
	return uri, nil
}

func addVal(values url.Values, key, val string) {
	if val != "" {
		values.Add(key, val)
	}
}

func encodeTags(values url.Values, prefix string, tags map[string]string) {
	if len(tags) == 0 {
		return
	}
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		// exclude pod-template-hash
		if k != "pod-template-hash" {
			values.Add("tag", fmt.Sprintf("%s%s:%s", prefix, k, tags[k]))
		}
	}
}
