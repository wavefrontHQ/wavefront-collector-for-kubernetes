package processors

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

type CumulativeDistributionConverter struct {
}

func (rc *CumulativeDistributionConverter) Name() string {
	return "cumulative distribution converter"
}

func (rc *CumulativeDistributionConverter) Process(batch *metrics.Batch) (*metrics.Batch, error) {
	for i, m := range batch.Metrics {
		distribution, ok := m.(*wf.Distribution)
		if !ok {
			continue
		}
		batch.Metrics[i] = distribution.ToFrequency()
	}
	return batch, nil
}

func NewCumulativeDistributionConverter() *CumulativeDistributionConverter {
	return &CumulativeDistributionConverter{}
}
