package processors

import (
	"sync"

	gometrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

func DuplicateHistogramCounter(name string) gometrics.Counter {
	return gometrics.GetOrRegisterCounter(
		reporting.EncodeKey("histograms.duplicates", map[string]string{"metricname": name}),
		gometrics.DefaultRegistry,
	)
}

type DistributionRateCalculator struct {
	lock              sync.Mutex
	prevDistributions map[wf.DistributionHash]*wf.Distribution
}

func (rc *DistributionRateCalculator) Name() string {
	return "distribution rate calculator"
}

func (rc *DistributionRateCalculator) Process(batch *metrics.Batch) (*metrics.Batch, error) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	seen := map[wf.DistributionHash]struct{}{}
	batch.Metrics = filterMapInPlace(func(metric wf.Metric) (wf.Metric, bool) {
		distribution, ok := metric.(*wf.Distribution)
		if !ok {
			return metric, true
		}
		if _, isDuplicate := seen[distribution.Key()]; isDuplicate {
			log.Warnf(
				"duplicate histogram series name=%s source=%s tags=%v",
				distribution.Name(), distribution.Source, distribution.Tags(),
			)
			DuplicateHistogramCounter(distribution.Name()).Inc(1)
			return nil, false
		}
		rate := distribution.Rate(rc.prevDistributions[distribution.Key()])
		rc.prevDistributions[distribution.Key()] = distribution.Clone()
		seen[distribution.Key()] = struct{}{}
		return rate, rate != nil
	}, batch.Metrics)
	return batch, nil
}

func filterMapInPlace(f func(wf.Metric) (wf.Metric, bool), es []wf.Metric) []wf.Metric {
	newEs := es[:0]
	for _, e := range es {
		newE, include := f(e)
		if include {
			newEs = append(newEs, newE)
		}
	}
	for i := range es[len(newEs):] {
		es[len(newEs)+i] = nil
	}
	return newEs
}

func NewDistributionRateCalculator() *DistributionRateCalculator {
	return &DistributionRateCalculator{prevDistributions: map[wf.DistributionHash]*wf.Distribution{}}
}
