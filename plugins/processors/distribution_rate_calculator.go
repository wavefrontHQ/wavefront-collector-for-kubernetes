package processors

import (
    log "github.com/sirupsen/logrus"
    "sync"
    "time"

    "github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

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
	n := len(batch.Metrics)
	newMetrics := make([]wf.Metric, 0, len(batch.Metrics))
	for i := 0; i < n; i++ {
		distribution, ok := batch.Metrics[i].(*wf.Distribution)
		if !ok {
			newMetrics = append(newMetrics, batch.Metrics[i])
			continue
		}
		rate := distribution.Rate(rc.prevDistributions[distribution.Key()])
		if rate != nil {
			newMetrics = append(newMetrics, rate)
		}
        if rc.prevDistributions[distribution.Key()] != nil && distribution.Timestamp.Sub(rc.prevDistributions[distribution.Key()].Timestamp) < time.Minute {
            log.Infof("Timestamp:: %+#v", distribution)
            log.Infof("TimestampDiff:: %s", distribution.Timestamp.Sub(rc.prevDistributions[distribution.Key()].Timestamp).String())
        }
        rc.prevDistributions[distribution.Key()] = distribution
	}
	batch.Metrics = newMetrics
	return batch, nil
}

func NewDistributionRateCalculator() *DistributionRateCalculator {
	return &DistributionRateCalculator{prevDistributions: map[wf.DistributionHash]*wf.Distribution{}}
}
