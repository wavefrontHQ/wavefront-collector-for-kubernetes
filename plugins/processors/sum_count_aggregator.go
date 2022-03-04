package processors

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	corev1 "k8s.io/api/core/v1"
)

type SumCountAggregator struct {
	name  string
	specs []SumCountAggregateSpec
}

type SumCountAggregateSpec struct {
	ResourceSumMetrics  []string
	ResourceCountMetric string
	ShouldAggregate     func(*metrics.Set) bool
	ExtractGroup        func(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set, error)
}

func NewSumCountAggregator(name string, specs []SumCountAggregateSpec) *SumCountAggregator {
	return &SumCountAggregator{name, specs}
}

func (a *SumCountAggregator) Name() string {
	return fmt.Sprintf("%s_aggregator", a.name)
}

func (a *SumCountAggregator) Process(batch *metrics.Batch) (*metrics.Batch, error) {
	for _, spec := range a.specs {
		for resourceKey, resourceSet := range batch.Sets {
			if !spec.ShouldAggregate(resourceSet) {
				continue
			}
			groupKey, groupSet, err := spec.ExtractGroup(batch, resourceKey, resourceSet)
			if err != nil {
				log.Errorf(err.Error())
				continue
			}
			if groupSet == nil {
				continue
			}
			aggregateCount(resourceSet, groupSet, spec.ResourceCountMetric)
			if err := aggregate(resourceSet, groupSet, spec.ResourceSumMetrics); err != nil {
				return nil, err
			}
			batch.Sets[groupKey] = groupSet
		}
	}
	return batch, nil
}

func isAggregatablePod(set *metrics.Set) bool {
	return isType(metrics.MetricSetTypePod)(set) && podTakesUpResources(set)
}

func isAggregatablePodContainer(set *metrics.Set) bool {
	return isType(metrics.MetricSetTypePodContainer)(set) && podContainerTakesUpResources(set)
}

func isType(matchType string) func(*metrics.Set) bool {
	return func(set *metrics.Set) bool {
		return set.Labels[metrics.LabelMetricSetType.Key] == matchType
	}
}

func podTakesUpResources(set *metrics.Set) bool {
	labels, _ := set.FindLabels(metrics.MetricPodPhase.Name)
	return labels["phase"] != string(corev1.PodSucceeded) && labels["phase"] != string(corev1.PodFailed)
}

func podContainerTakesUpResources(set *metrics.Set) bool {
	labels, _ := set.FindLabels(metrics.MetricContainerStatus.Name)
	return labels["state"] != "terminated"
}
