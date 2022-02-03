// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package wavefront

import (
	"strings"
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/stretchr/testify/assert"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

func NewTestWavefrontSink() *wavefrontSink {
	return &wavefrontSink{
		WavefrontClient: NewTestSender(),
		ClusterName:     "testCluster",
	}
}

func TestStoreTimeseriesEmptyInput(t *testing.T) {
	fakeSink := NewTestWavefrontSink()
	db := metrics.DataBatch{}
	fakeSink.ExportData(&db)
	assert.Equal(t, 0, len(getMetrics(fakeSink)))
}

func TestName(t *testing.T) {
	fakeSink := NewTestWavefrontSink()
	name := fakeSink.Name()
	assert.Equal(t, name, "wavefront_sink")
}

func TestCreateWavefrontSinkWithNoEmptyInputs(t *testing.T) {
	cfg := configuration.WavefrontSinkConfig{
		ProxyAddress: "wavefront-proxy:2878",
		ClusterName:  "testCluster",
		TestMode:     true,
		Transforms: configuration.Transforms{
			Prefix: "testPrefix",
		},
	}
	sink, err := NewWavefrontSink(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, sink)
	wfSink, ok := sink.(*wavefrontSink)
	assert.Equal(t, true, ok)
	assert.NotNil(t, wfSink.WavefrontClient)
	assert.Equal(t, "testCluster", wfSink.ClusterName)
	assert.Equal(t, "testPrefix", wfSink.Prefix)
}

func TestPrefix(t *testing.T) {
	cfg := configuration.WavefrontSinkConfig{
		ProxyAddress: "wavefront-proxy:2878",
		TestMode:     true,
		Transforms: configuration.Transforms{
			Prefix: "test.",
		},
	}
	sink, err := NewWavefrontSink(cfg)
	assert.NoError(t, err)

	db := metrics.DataBatch{
		Points: []*wf.Point{
			wf.NewPoint("cpu.idle", 1.0, 0, "fakeSource", nil),
		},
	}
	sink.ExportData(&db)
	assert.True(t, strings.Contains(getMetrics(sink), "test.cpu.idle"))
}

func TestNilPointDataBatch(t *testing.T) {
	cfg := configuration.WavefrontSinkConfig{
		ProxyAddress: "wavefront-proxy:2878",
		TestMode:     true,
		Transforms: configuration.Transforms{
			Prefix: "test.",
		},
	}
	sink, err := NewWavefrontSink(cfg)
	assert.NoError(t, err)

	db := metrics.DataBatch{
		Points: []*wf.Point{
			wf.NewPoint("cpu.idle", 1.0, 0, "fakeSource", nil),
			nil,
		},
	}
	sink.ExportData(&db)
	assert.True(t, strings.Contains(getMetrics(sink), "test.cpu.idle"))
}

func TestCleansTagsBeforeSending(t *testing.T) {
	cfg := configuration.WavefrontSinkConfig{
		ProxyAddress: "wavefront-proxy:2878",
		TestMode:     true,
		Transforms: configuration.Transforms{
			Prefix: "test.",
		},
	}
	sink, err := NewWavefrontSink(cfg)
	assert.NoError(t, err)

	db := metrics.DataBatch{
		Points: []*wf.Point{
			wf.NewPoint(
				"cpu.idle",
				1.0,
				0,
				"fakeSource",
				map[string]string{"emptyTag": ""},
			),
		},
	}
	sink.ExportData(&db)
	assert.NotContains(t, getMetrics(sink), "emptyTag")
}

func getMetrics(sink WavefrontSink) string {
	return strings.TrimSpace(sink.(*wavefrontSink).WavefrontClient.(*TestSender).GetReceivedLines())
}
