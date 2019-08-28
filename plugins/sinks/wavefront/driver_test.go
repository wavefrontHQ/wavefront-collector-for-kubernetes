// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package wavefront

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

func NewFakeWavefrontSink() *wavefrontSink {
	return &wavefrontSink{
		testMode:    true,
		ClusterName: "testCluster",
	}
}

func TestStoreTimeseriesEmptyInput(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	db := metrics.DataBatch{}
	fakeSink.ExportData(&db)
	assert.Equal(t, 0, len(fakeSink.testReceivedLines))
}

func TestName(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	name := fakeSink.Name()
	assert.Equal(t, name, "wavefront_sink")
}

func TestCreateWavefrontSinkWithNoEmptyInputs(t *testing.T) {
	cfg := configuration.WavefrontSinkConfig{
		ProxyAddress: "wavefront-proxy:2878",
		ClusterName:  "testCluster",
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
