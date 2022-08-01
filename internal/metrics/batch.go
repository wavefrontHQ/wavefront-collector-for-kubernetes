package metrics

import (
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

// Batch contains sets of metrics tied to specific k8s resources and other more general wavefront points
type Batch struct {
	Timestamp time.Time
	Sets      map[ResourceKey]*Set
	Metrics   []wf.Metric
}

func (b *Batch) Points() int {
	total := 0
	for _, set := range b.Sets {
		total += set.Points()
	}
	for _, metric := range b.Metrics {
		total += metric.Points()
	}
	return total
}
