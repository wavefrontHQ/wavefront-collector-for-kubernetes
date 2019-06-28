package wavefront

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, name, "Wavefront Sink")
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
}
