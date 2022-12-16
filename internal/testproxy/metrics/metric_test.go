package metrics_test

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/metrics"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMetric(t *testing.T) {
	t.Run("can parse histograms", func(t *testing.T) {
		metric, err := metrics.ParseMetric("!M 1493773500 #20 30 #10 5 request.latency source=\"appServer1\" region=\"us-west\"")
		assert.Nil(t, metric)
		assert.NoError(t, err)
	})

	t.Run("can parse metrics", func(t *testing.T) {
		metric, err := metrics.ParseMetric("system.cpu.loadavg.1m 0.03 1382754475 source=\"test1.wavefront.com\"")
		assert.NoError(t, err)
		assert.Equal(t, "system.cpu.loadavg.1m", metric.Name)
		assert.Equal(t, "0.03", metric.Value)
		assert.Equal(t, "1382754475", metric.Timestamp)
		assert.Equal(t, map[string]string{"source": "test1.wavefront.com"}, metric.Tags)
	})
}
