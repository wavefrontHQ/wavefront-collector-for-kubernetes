package senders

import (
	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

type noOpSender struct {
}

var (
	defaultNoopClient Sender = &noOpSender{}
)

// NewWavefrontNoOpClient returns a Wavefront Client instance for which all operations are no-ops.
func NewWavefrontNoOpClient() (Sender, error) {
	return defaultNoopClient, nil
}

func (sender *noOpSender) private() {
}

func (sender *noOpSender) Start() {
	// no-op
}

func (sender *noOpSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	return nil
}

func (sender *noOpSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	return nil
}

func (sender *noOpSender) SendDistribution(name string, centroids []histogram.Centroid,
	hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	return nil
}

func (sender *noOpSender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string,
	parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	return nil
}

func (sender *noOpSender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	return nil
}

func (sender *noOpSender) Close() {
	// no-op
}

func (sender *noOpSender) Flush() error {
	return nil
}

func (sender *noOpSender) GetFailureCount() int64 {
	return 0
}
