package wavefront

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
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

func NewFakeWavefrontSink() *wavefrontSink {
	return &wavefrontSink{
		testMode:          true,
		ClusterName:       "testCluster",
		IncludeLabels:     false,
		IncludeContainers: true,
	}
}

func TestStoreTimeseriesEmptyInput(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	db := metrics.DataBatch{}
	fakeSink.ExportData(&db)
	assert.Equal(t, 0, len(fakeSink.testReceivedLines))
}

func TestStoreTimeseriesMultipleTimeseriesInput(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	batch := generateFakeBatch()
	fakeSink.ExportData(batch)
	assert.Equal(t, len(batch.MetricSets), len(fakeSink.testReceivedLines))
}
func TestName(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	name := fakeSink.Name()
	assert.Equal(t, name, "Wavefront Sink")
}

func TestNewMetricName(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	batch := generateFakeBatch()
	fakeSink.ExportData(batch)
	name := "cpu/usage"
	mtype := "pod_container"
	newName := fakeSink.cleanMetricName(mtype, name)
	assert.Equal(t, "pod_container.cpu.usage", newName)
}

func TestValidateLines(t *testing.T) {
	fakeSink := NewFakeWavefrontSink()
	batch := generateFakeBatch()
	fakeSink.ExportData(batch)

	//validate each line received from the fake batch
	for _, line := range fakeSink.testReceivedLines {
		parts := strings.Split(strings.TrimSpace(line), " ")

		//second part should always be the numeric metric value
		_, err := strconv.ParseFloat(parts[1], 64)
		assert.NoError(t, err)

		//third part should always be the epoch timestamp (a count of seconds)
		_, err = strconv.ParseInt(parts[2], 0, 64)
		assert.NoError(t, err)

		//the fourth part should be the source tag
		isSourceTag := strings.HasPrefix(parts[3], "source=")
		assert.True(t, isSourceTag)

		//all remaining parts are tags and must be key value pairs (containing "=")
		tags := parts[4:]
		fmt.Println(tags)
		for _, v := range tags {
			assert.True(t, strings.Contains(v, "="))
		}
	}
}

func TestCreateWavefrontSinkWithNoEmptyInputs(t *testing.T) {
	fakeUrl := "?proxyAddress=wavefront-proxy:2878&clusterName=testCluster&prefix=testPrefix&includeLabels=true&includeContainers=true"
	uri, _ := url.Parse(fakeUrl)
	sink, err := NewWavefrontSink(uri)
	assert.NoError(t, err)
	assert.NotNil(t, sink)
	wfSink, ok := sink.(*wavefrontSink)
	assert.Equal(t, true, ok)
	assert.NotNil(t, wfSink.WavefrontClient)
	assert.Equal(t, "testCluster", wfSink.ClusterName)
	assert.Equal(t, "testPrefix", wfSink.Prefix)
	assert.Equal(t, true, wfSink.IncludeLabels)
	assert.Equal(t, true, wfSink.IncludeContainers)
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
