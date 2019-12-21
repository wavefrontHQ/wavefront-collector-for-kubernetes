package kstate

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	gometrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"

	"k8s.io/apimachinery/pkg/labels"
	v1lister "k8s.io/client-go/listers/core/v1"
)

var (
	collectErrors   gometrics.Counter
	filteredPoints  gometrics.Counter
	collectedPoints gometrics.Counter
)

func init() {
	pt := map[string]string{"type": "kubernetes.state"}
	collectedPoints = gometrics.GetOrRegisterCounter(reporting.EncodeKey("source.points.collected", pt), gometrics.DefaultRegistry)
	filteredPoints = gometrics.GetOrRegisterCounter(reporting.EncodeKey("source.points.filtered", pt), gometrics.DefaultRegistry)
	collectErrors = gometrics.GetOrRegisterCounter(reporting.EncodeKey("source.collect.errors", pt), gometrics.DefaultRegistry)
}

type stateMetricsSource struct {
	podLister v1lister.PodLister
	prefix    string
	source    string
	tags      map[string]string
	filters   filter.Filter
}

func NewStateMetricsSource(podLister v1lister.PodLister, transforms configuration.Transforms) (metrics.MetricsSource, error) {
	return &stateMetricsSource{
		podLister: podLister,
		prefix:    transforms.Prefix,
		source:    getDefault(util.GetNodeName(), transforms.Source),
		tags:      transforms.Tags,
		filters:   filter.FromConfig(transforms.Filters),
	}, nil
}

func getDefault(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func (src *stateMetricsSource) Name() string {
	return "kstate_source"
}

func (src *stateMetricsSource) ScrapeMetrics() (*metrics.DataBatch, error) {
	result := &metrics.DataBatch{
		Timestamp: time.Now(),
	}
	points, err := src.buildPodStatus()
	if err != nil {
		collectErrors.Inc(1)
		return result, err
	}
	result.MetricPoints = points
	collectedPoints.Inc(int64(len(points)))
	return result, nil
}

func (src *stateMetricsSource) buildPodStatus() ([]*metrics.MetricPoint, error) {
	pods, err := src.podLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	var points []*metrics.MetricPoint
	for _, pod := range pods {
		tags := map[string]string{
			"pod_name":       pod.Name,
			"namespace_name": pod.Namespace,
		}
		points = src.filter(points, buildPodPhase(pod, src.prefix, src.source, tags, now))
		points = src.filterAppend(points, buildContainerStatuses(pod.Status.ContainerStatuses, src.prefix+"pod_container.", src.source, tags, now))
		points = src.filterAppend(points, buildContainerStatuses(pod.Status.InitContainerStatuses, src.prefix+"pod_init_container.", src.source, tags, now))
	}
	return points, nil
}

func (src *stateMetricsSource) filter(slice []*metrics.MetricPoint, point *metrics.MetricPoint) []*metrics.MetricPoint {
	if src.filters == nil || src.filters.Match(point.Metric, point.Tags) {
		return append(slice, point)
	}
	filteredPoints.Inc(1)
	log.Tracef("dropping metric: %s", point.Metric)
	return slice
}

func (src *stateMetricsSource) filterAppend(slice []*metrics.MetricPoint, points []*metrics.MetricPoint) []*metrics.MetricPoint {
	for _, point := range points {
		slice = src.filter(slice, point)
	}
	return slice
}

type stateProvider struct {
	metrics.DefaultMetricsSourceProvider
	sources []metrics.MetricsSource
}

func (p *stateProvider) GetMetricsSources() []metrics.MetricsSource {
	return p.sources
}

func (p *stateProvider) Name() string {
	return providerName
}

const providerName = "kubernetes_state_metrics_provider"

func NewStateProvider(cfg configuration.KubernetesStateSourceConfig) (metrics.MetricsSourceProvider, error) {
	podLister, err := util.GetPodLister(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating podLister: %v", err)
	}

	var sources []metrics.MetricsSource
	metricsSource, err := NewStateMetricsSource(podLister, cfg.Transforms)
	if err == nil {
		sources = append(sources, metricsSource)
	} else {
		return nil, fmt.Errorf("error creating source: %v", err)
	}

	return &stateProvider{
		sources: sources,
	}, nil
}
