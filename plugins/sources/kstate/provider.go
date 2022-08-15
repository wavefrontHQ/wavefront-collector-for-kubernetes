// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"fmt"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	"github.com/wavefronthq/go-metrics-wavefront/reporting"

	gometrics "github.com/rcrowley/go-metrics"
)

type resourceHandler func(interface{}, configuration.Transforms) []wf.Metric

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

func NewStateMetricsSource(lister *lister, transforms configuration.Transforms) (metrics.Source, error) {
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
	funcs[replicationControllers] = pointsForReplicationController
	funcs[statefulSets] = pointsForStatefulSet
	funcs[horizontalPodAutoscalers] = pointsForHPA
	funcs[nodes] = pointsForNode
	funcs[nonRunningPods] = pointsForNonRunningPods

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

func (src *stateMetricsSource) AutoDiscovered() bool {
	return false
}

func (src *stateMetricsSource) Name() string {
	return "kstate_source"
}

func (src *stateMetricsSource) Cleanup() {}

func (src *stateMetricsSource) Scrape() (*metrics.Batch, error) {
	result := &metrics.Batch{
		Timestamp: time.Now(),
	}

	var points []wf.Metric
	for resType := range src.funcs {
		for _, point := range src.pointsForResource(resType) {
			points = wf.FilterAppend(src.filters, src.fps, points, point)
		}
	}
	result.Metrics = points
	src.pps.Inc(int64(len(result.Metrics)))
	return result, nil
}

func (src *stateMetricsSource) pointsForResource(resType string) []wf.Metric {
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

	var points []wf.Metric
	for _, item := range items {
		points = append(points, f(item, src.transforms)...)
	}
	return points
}

type stateProvider struct {
	metrics.DefaultSourceProvider
	sources []metrics.Source
}

func (p *stateProvider) GetMetricsSources() []metrics.Source {
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

func NewStateProvider(cfg configuration.KubernetesStateSourceConfig) (metrics.SourceProvider, error) {
	if !util.ScrapeCluster() {
		return &stateProvider{}, nil
	}

	if cfg.KubeClient == nil {
		return nil, fmt.Errorf("kubeclient not initialized")
	}

	var sources []metrics.Source
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
