package kstate

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	"k8s.io/apimachinery/pkg/labels"
	v1lister "k8s.io/client-go/listers/core/v1"
)

type stateMetricsSource struct {
	podLister v1lister.PodLister
	prefix    string
	source    string
	tags      map[string]string
	filters   filter.Filter
	//pps        gometrics.Counter
	//eps        gometrics.Counter
}

func NewStateMetricsSource(podLister v1lister.PodLister, transforms configuration.Transforms) (metrics.MetricsSource, error) {
	//TODO: emit internal metrics
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
	result.MetricPoints = src.buildPodStatus()
	return result, nil
}

func (src *stateMetricsSource) buildPodStatus() []*metrics.MetricPoint {
	pods, err := src.podLister.List(labels.Everything())
	if err != nil {
		return nil
	}

	log.Debugf("total pods retrieved: %d", len(pods))

	if len(pods) > 0 {
		log.Debugf("first pod info: %v", pods[0].Status)
	}

	now := time.Now().Unix()
	var points []*metrics.MetricPoint
	for _, pod := range pods {
		tags := map[string]string{
			"pod_name":       pod.Name,
			"namespace_name": pod.Namespace,
		}
		points = append(points, buildPodPhase(pod, src.prefix, src.source, tags, now))
	}
	return points
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

const providerName = "kstate_metrics_provider"

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
