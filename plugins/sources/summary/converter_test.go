// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package summary

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

var (
	fakeNodeIp  = "192.168.1.23"
	fakePodName = "redis-test"
	fakePodUid  = "redis-test-uid"
	fakeLabel   = map[string]string{
		"name":                   "redis",
		"io.kubernetes.pod.name": "default/redis-test",
		"pod_id":                 fakePodUid,
		"namespace_name":         "default",
		"pod_name":               fakePodName,
		"container_name":         "redis",
		"container_base_image":   "kubernetes/redis:v1",
		"namespace_id":           "namespace-test-uid",
		"host_id":                fakeNodeIp,
		"hostname":               fakeNodeIp,
	}
)

func TestNewMetricName(t *testing.T) {
	converter := fakeWavefrontConverter(t, configuration.SummaySourceConfig{})
	name := "cpu/usage"
	mtype := "pod_container"
	newName := converter.(*pointConverter).cleanMetricName(mtype, name)
	assert.Equal(t, "kubernetes.pod_container.cpu.usage", newName)
}

func TestStoreTimeseriesMultipleTimeseriesInput(t *testing.T) {
	fakeConverter := fakeWavefrontConverter(t, configuration.SummaySourceConfig{})
	batch := generateFakeBatch()
	count := len(batch.MetricSets)
	data, err := fakeConverter.Process(batch)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, count, len(data.MetricPoints))
}

func TestFiltering(t *testing.T) {
	fakeConverter := fakeWavefrontConverter(t, configuration.SummaySourceConfig{
		Transforms: configuration.Transforms{
			Filters: filter.Config{
				MetricWhitelist: []string{"kubernetes*cpu*"},
			},
		},
	})
	batch := generateFakeBatch()
	data, err := fakeConverter.Process(batch)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 8, len(data.MetricSets))
	assert.Equal(t, 2, len(data.MetricPoints))

	fakeConverter = fakeWavefrontConverter(t, configuration.SummaySourceConfig{
		Transforms: configuration.Transforms{
			Filters: filter.Config{
				MetricBlacklist: []string{"kubernetes*cpu*"},
			},
		},
	})
	batch = generateFakeBatch()
	data, err = fakeConverter.Process(batch)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 8, len(data.MetricSets))
	assert.Equal(t, 6, len(data.MetricPoints))
}

func TestPrefix(t *testing.T) {
	fakeConverter := fakeWavefrontConverter(t, configuration.SummaySourceConfig{
		Transforms: configuration.Transforms{Prefix: "k8s."},
	})
	name := "cpu/usage"
	mtype := "pod"
	newName := fakeConverter.(*pointConverter).cleanMetricName(mtype, name)
	assert.Equal(t, "k8s.pod.cpu.usage", newName)
}

func generateFakeBatch() *metrics.DataBatch {
	batch := metrics.DataBatch{
		Timestamp:  time.Now(),
		MetricSets: map[string]*metrics.MetricSet{},
	}

	batch.MetricSets["m1"] = generateMetricSet("cpu/limit", metrics.MetricGauge, 1000)
	batch.MetricSets["m2"] = generateMetricSet("cpu/usage", metrics.MetricCumulative, 43363664)
	batch.MetricSets["m3"] = generateMetricSet("filesystem/limit", metrics.MetricGauge, 42241163264)
	batch.MetricSets["m4"] = generateMetricSet("filesystem/usage", metrics.MetricGauge, 32768)
	batch.MetricSets["m5"] = generateMetricSet("memory/limit", metrics.MetricGauge, -1)
	batch.MetricSets["m6"] = generateMetricSet("memory/usage", metrics.MetricGauge, 487424)
	batch.MetricSets["m7"] = generateMetricSet("memory/working_set", metrics.MetricGauge, 491520)
	batch.MetricSets["m8"] = generateMetricSet("uptime", metrics.MetricCumulative, 910823)
	return &batch
}

func generateMetricSet(name string, metricType metrics.MetricType, value int64) *metrics.MetricSet {
	set := &metrics.MetricSet{
		Labels: fakeLabel,
		MetricValues: map[string]metrics.MetricValue{
			name: {
				MetricType: metricType,
				ValueType:  metrics.ValueInt64,
				IntValue:   value,
			},
		},
	}
	return set
}

func fakeWavefrontConverter(t *testing.T, cfg configuration.SummaySourceConfig) metrics.DataProcessor {
	converter, err := NewPointConverter(cfg, "k8s-cluster")
	if err != nil {
		t.Error(err)
	}
	return converter
}
