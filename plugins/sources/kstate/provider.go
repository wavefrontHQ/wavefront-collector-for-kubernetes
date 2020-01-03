// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	"github.com/wavefronthq/go-metrics-wavefront/reporting"

	gometrics "github.com/rcrowley/go-metrics"
)

type resourceHandler func(interface{}, configuration.Transforms) []*metrics.MetricPoint

type stateMetricsSource struct {
	lister     *lister
	transforms configuration.Transforms
	source     string
	filters    filter.Filter
	funcs      map[string]resourceHandler

	pps gometrics.Counter
	eps gometrics.Counter
	fps gometrics.Counter
}

func NewStateMetricsSource(lister *lister, transforms configuration.Transforms) (metrics.MetricsSource, error) {
	pt := map[string]string{"type": "kubernetes.state"}
	ppsKey := reporting.EncodeKey("source.points.collected", pt)
	epsKey := reporting.EncodeKey("source.collect.errors", pt)
	fpsKey := reporting.EncodeKey("source.points.filtered", pt)

	transforms.Source = getDefault(util.GetNodeName(), transforms.Source)
	transforms.Prefix = getDefault(transforms.Prefix, "kubernetes.")

	funcs := make(map[string]resourceHandler)
	funcs[jobs] = pointsForJob
	funcs[cronJobs] = pointsForCronJob
	funcs[daemonSets] = pointsForDaemonSet
	funcs[deployments] = pointsForDeployment
	funcs[replicaSets] = pointsForReplicaSet
	funcs[statefulSets] = pointsForStatefulSet
	funcs[horizontalPodAutoscalers] = pointsForHPA

	return &stateMetricsSource{
		lister:     lister,
		transforms: transforms,
		filters:    filter.FromConfig(transforms.Filters),
		funcs:      funcs,
		pps:        gometrics.GetOrRegisterCounter(ppsKey, gometrics.DefaultRegistry),
		eps:        gometrics.GetOrRegisterCounter(epsKey, gometrics.DefaultRegistry),
		fps:        gometrics.GetOrRegisterCounter(fpsKey, gometrics.DefaultRegistry),
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

	var points []*metrics.MetricPoint
	for resType := range src.funcs {
		points = append(points, src.pointsForResource(resType)...)
	}

	n := 0
	for _, point := range points {
		if src.keep(point.Metric, point.Tags) {
			// in-place filtering
			points[n] = point
			n++
		}
	}
	result.MetricPoints = points[:n]

	src.pps.Inc(int64(len(result.MetricPoints)))
	return result, nil
}

func (src *stateMetricsSource) pointsForResource(resType string) []*metrics.MetricPoint {
	items, err := src.lister.List(resType)
	if err != nil {
		log.Errorf("error listing %s: %v", resType, err)
		return nil
	}

	if len(items) == 0 {
		return nil
	}

	f, ok := src.funcs[resType]
	if !ok {
		return nil
	}

	var points []*metrics.MetricPoint
	for _, item := range items {
		points = append(points, f(item, src.transforms)...)
	}
	return points
}

func (src *stateMetricsSource) keep(name string, tags map[string]string) bool {
	if src.filters == nil || src.filters.Match(name, tags) {
		return true
	}
	src.fps.Inc(1)
	log.Tracef("dropping metric: %s", name)
	return false
}

type stateProvider struct {
	metrics.DefaultMetricsSourceProvider
	sources []metrics.MetricsSource
}

func (p *stateProvider) GetMetricsSources() []metrics.MetricsSource {
	if !leadership.Leading() {
		log.Infof("not scraping sources from: %s. current leader: %s", providerName, leadership.Leader())
		return nil
	}
	return p.sources
}

func (p *stateProvider) Name() string {
	return providerName
}

const providerName = "kstate_metrics_provider"

func NewStateProvider(cfg configuration.KubernetesStateSourceConfig) (metrics.MetricsSourceProvider, error) {
	if cfg.KubeClient == nil {
		return nil, fmt.Errorf("kubeclient not initialized")
	}

	var sources []metrics.MetricsSource
	metricsSource, err := NewStateMetricsSource(newLister(cfg.KubeClient), cfg.Transforms)
	if err == nil {
		sources = append(sources, metricsSource)
	} else {
		return nil, fmt.Errorf("error creating source: %v", err)
	}

	return &stateProvider{
		sources: sources,
	}, nil
}
