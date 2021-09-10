// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"
)

func TestMetricEmptyFile(t *testing.T) {
	byteReader := bytes.NewReader([]byte{})
	reader := prometheus.NewMetricReader(byteReader)
	assertEmptyMetricReader(t, reader)
}

var metricOne string = `# HELP ack_level_update ack_level_update counter
# TYPE ack_level_update counter
ack_level_update{operation="TimerActiveQueueProcessor",type="history"} 1.599204e+06
ack_level_update{operation="TransferActiveQueueProcessor",type="history"} 1.599186e+06
`

var metricTwo string = `# HELP acquire_shards_count acquire_shards_count counter
# TYPE acquire_shards_count counter
acquire_shards_count{operation="ShardController",type="history"} 2904
`

func TestMetric(t *testing.T) {
	file := append([]byte(metricOne), []byte(metricTwo)...)

	byteReader := bytes.NewReader(file)
	reader := prometheus.NewMetricReader(byteReader)

	assert.False(t, reader.Done())
	assert.Equal(t, metricOne, string(reader.Read()))

	assert.False(t, reader.Done())
	assert.Equal(t, metricTwo, string(reader.Read()))

	assertEmptyMetricReader(t, reader)
}

func TestMetricBlankLines(t *testing.T) {
	var spaced string = `
# HELP ack_level_update ack_level_update counter
  
# TYPE ack_level_update counter
  
ack_level_update{operation="TimerActiveQueueProcessor",type="history"} 1.599204e+06

ack_level_update{operation="TransferActiveQueueProcessor",type="history"} 1.599186e+06

`

	file := append([]byte(spaced), []byte(metricTwo)...)

	byteReader := bytes.NewReader(file)
	reader := prometheus.NewMetricReader(byteReader)

	assert.False(t, reader.Done())
	assert.Equal(t, spaced, string(reader.Read()))

	assert.False(t, reader.Done())
	assert.Equal(t, metricTwo, string(reader.Read()))

	assertEmptyMetricReader(t, reader)
}

func TestMetricLeadingWhitespace(t *testing.T) {
	leading := `  # HELP ack_level_update ack_level_update counter
	# TYPE ack_level_update counter
  ack_level_update{operation="TimerActiveQueueProcessor",type="history"} 1.599204e+06
	ack_level_update{operation="TransferActiveQueueProcessor",type="history"} 1.599186e+06
`

	file := append([]byte(leading), []byte(metricTwo)...)

	byteReader := bytes.NewReader(file)
	reader := prometheus.NewMetricReader(byteReader)

	assert.False(t, reader.Done())
	assert.Equal(t, leading, string(reader.Read()))

	assert.False(t, reader.Done())
	assert.Equal(t, metricTwo, string(reader.Read()))

	assertEmptyMetricReader(t, reader)
}

func assertEmptyMetricReader(t *testing.T, reader *prometheus.MetricReader) {
	assert.True(t, reader.Done())
	assert.Empty(t, reader.Read())
}
