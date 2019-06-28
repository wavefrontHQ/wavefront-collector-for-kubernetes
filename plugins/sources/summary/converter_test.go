package summary

import (
	"net/url"
	"testing"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"github.com/stretchr/testify/assert"
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
	converter := fakeWavefrontConverter(t, "?")
	name := "cpu/usage"
	mtype := "pod_container"
	newName := converter.(*pointConverter).cleanMetricName(mtype, name)
	assert.Equal(t, "kubernetes.pod_container.cpu.usage", newName)
}

func TestStoreTimeseriesMultipleTimeseriesInput(t *testing.T) {
	fakeConverter := fakeWavefrontConverter(t, "?")
	batch := generateFakeBatch()
	count := len(batch.MetricSets)
	data, err := fakeConverter.Process(batch)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(data.MetricSets), 0)
	assert.Equal(t, count, len(data.MetricPoints))
}

func TestFiltering(t *testing.T) {
	fakeConverter := fakeWavefrontConverter(t, "?metricWhitelist=kubernetes*cpu*")
	batch := generateFakeBatch()
	data, err := fakeConverter.Process(batch)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(data.MetricSets), 0)
	assert.Equal(t, 2, len(data.MetricPoints))

	fakeConverter = fakeWavefrontConverter(t, "?metricBlacklist=kubernetes*cpu*")
	batch = generateFakeBatch()
	data, err = fakeConverter.Process(batch)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(data.MetricSets), 0)
	assert.Equal(t, 6, len(data.MetricPoints))
}

func TestPrefix(t *testing.T) {
	fakeConverter := fakeWavefrontConverter(t, "?prefix=k8s.")
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

func fakeWavefrontConverter(t *testing.T, uri string) metrics.DataProcessor {
	u, err := url.Parse(uri)
	if err != nil {
		t.Error("error creating url")
	}

	converter, err := NewPointConverter(u, "k8s-cluster")
	if err != nil {
		t.Error(err)
	}
	return converter
}
