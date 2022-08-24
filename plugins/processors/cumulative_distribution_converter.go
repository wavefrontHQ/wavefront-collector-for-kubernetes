package processors

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/slicextra"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

type CumulativeDistributionConverter struct {
}

func (rc *CumulativeDistributionConverter) Name() string {
	return "cumulative distribution converter"
}

func (rc *CumulativeDistributionConverter) Process(batch *metrics.Batch) (*metrics.Batch, error) {
	batch.Metrics = slicextra.MapInPlace[wf.Metric](func(metric wf.Metric) wf.Metric {
		distribution, ok := metric.(*wf.Distribution)
		if !ok {
			return metric
		}
		return distribution.ToFrequency()
	}, batch.Metrics)
	return batch, nil
}

func NewCumulativeDistributionConverter() *CumulativeDistributionConverter {
	return &CumulativeDistributionConverter{}
}
