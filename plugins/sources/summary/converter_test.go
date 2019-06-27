package summary

import "testing"

func TestNewMetricName(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	batch := generateFakeBatch()
	fakeSink.ExportData(batch)
	name := "cpu/usage"
	mtype := "pod_container"
	newName := fakeSink.cleanMetricName(mtype, name)
	assert.Equal(t, "pod_container.cpu.usage", newName)
}
