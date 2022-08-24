package processors

import (
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
	"sync"
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
		} else if distribution.Name() == "kubernetes.controlplane.etcd.request.duration.seconds" {
			log.Infof("NilDistribution:: %+#v", distribution)
			log.Infof("NilDistribution:: %+#v", rc.prevDistributions[distribution.Key()])
		}
		rc.prevDistributions[distribution.Key()] = distribution.Clone() // TODO TDD
	}
	batch.Metrics = newMetrics
	return batch, nil
}

func NewDistributionRateCalculator() *DistributionRateCalculator {
	return &DistributionRateCalculator{prevDistributions: map[wf.DistributionHash]*wf.Distribution{}}
}
