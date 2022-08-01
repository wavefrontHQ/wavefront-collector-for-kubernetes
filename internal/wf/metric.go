package wf

import "github.com/wavefronthq/wavefront-sdk-go/histogram"

type Metric interface {
	Name() string
	Tags() map[string]string
	FilterTags(pred func(string) bool)
	OverrideTag(name, value string)
	AddTags(tags map[string]string)
	SetSource(source string)
	Send(to Sender) error
	Points() int
}

type Sender interface {
	SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error
	SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error
}
