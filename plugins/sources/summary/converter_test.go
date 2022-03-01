// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package summary

import (
	"testing"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"

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
	converter := fakeWavefrontConverter(t, configuration.SummarySourceConfig{})
	name := "cpu/usage"
	mtype := "pod_container"
	newName := converter.(*pointConverter).cleanMetricName(mtype, name)
	assert.Equal(t, "kubernetes.pod_container.cpu.usage", newName)
}

func TestStoreTimeseriesMultipleTimeseriesInput(t *testing.T) {
	fakeConverter := fakeWavefrontConverter(t, configuration.SummarySourceConfig{})
	batch := generateFakeBatch()
	count := len(batch.Sets)
	data, err := fakeConverter.Process(batch)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, count, len(data.Points))
}

func TestFiltering(t *testing.T) {
	fakeConverter := fakeWavefrontConverter(t, configuration.SummarySourceConfig{
		Transforms: configuration.Transforms{
			Filters: filter.Config{
				MetricAllowList: []string{"kubernetes*cpu*"},
			},
		},
	})
	batch := generateFakeBatch()
	data, err := fakeConverter.Process(batch)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 8, len(data.Sets))
	assert.Equal(t, 2, len(data.Points))

	fakeConverter = fakeWavefrontConverter(t, configuration.SummarySourceConfig{
		Transforms: configuration.Transforms{
			Filters: filter.Config{
				MetricDenyList: []string{"kubernetes*cpu*"},
			},
		},
	})
	batch = generateFakeBatch()
	data, err = fakeConverter.Process(batch)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 8, len(data.Sets))
	assert.Equal(t, 6, len(data.Points))
}

func TestPrefix(t *testing.T) {
	fakeConverter := fakeWavefrontConverter(t, configuration.SummarySourceConfig{
		Transforms: configuration.Transforms{Prefix: "k8s."},
	})
	name := "cpu/usage"
	mtype := "pod"
	newName := fakeConverter.(*pointConverter).cleanMetricName(mtype, name)
	assert.Equal(t, "k8s.pod.cpu.usage", newName)
}

func generateFakeBatch() *metrics.Batch {
	batch := metrics.Batch{
		Timestamp: time.Now(),
		Sets:      map[metrics.ResourceKey]*metrics.Set{},
	}

	batch.Sets["m1"] = generateMetricSet("cpu/limit", metrics.Gauge, 1000)
	batch.Sets["m2"] = generateMetricSet("cpu/usage", metrics.Cumulative, 43363664)
	batch.Sets["m3"] = generateMetricSet("filesystem/limit", metrics.Gauge, 42241163264)
	batch.Sets["m4"] = generateMetricSet("filesystem/usage", metrics.Gauge, 32768)
	batch.Sets["m5"] = generateMetricSet("memory/limit", metrics.Gauge, -1)
	batch.Sets["m6"] = generateMetricSet("memory/usage", metrics.Gauge, 487424)
	batch.Sets["m7"] = generateMetricSet("memory/working_set", metrics.Gauge, 491520)
	batch.Sets["m8"] = generateMetricSet("uptime", metrics.Cumulative, 910823)
	return &batch
}

func generateMetricSet(name string, metricType metrics.Type, value int64) *metrics.Set {
	set := &metrics.Set{
		Labels: fakeLabel,
		Values: map[string]metrics.Value{
			name: {
				ValueType: metrics.ValueInt64,
				IntValue:  value,
			},
		},
	}
	return set
}

func fakeWavefrontConverter(t *testing.T, cfg configuration.SummarySourceConfig) metrics.Processor {
	converter, err := NewPointConverter(cfg, "k8s-cluster")
	if err != nil {
		t.Error(err)
	}
	return converter
}
