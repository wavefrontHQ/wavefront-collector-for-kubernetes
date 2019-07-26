// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processors

import (
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	kube_config "github.com/wavefronthq/wavefront-kubernetes-collector/internal/kubernetes"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	kube_api "k8s.io/api/core/v1"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type NamespaceBasedEnricher struct {
	store cache.Store
}

func (this *NamespaceBasedEnricher) Name() string {
	return "namespace_based_enricher"
}

func (this *NamespaceBasedEnricher) Process(batch *metrics.DataBatch) (*metrics.DataBatch, error) {
	for _, ms := range batch.MetricSets {
		this.addNamespaceInfo(ms)
	}
	return batch, nil
}

// Adds UID to all namespaced elements.
func (this *NamespaceBasedEnricher) addNamespaceInfo(metricSet *metrics.MetricSet) {
	metricSetType, found := metricSet.Labels[metrics.LabelMetricSetType.Key]
	if !found {
		return
	}
	if metricSetType != metrics.MetricSetTypePodContainer &&
		metricSetType != metrics.MetricSetTypePod &&
		metricSetType != metrics.MetricSetTypeNamespace {
		return
	}

	namespaceName, found := metricSet.Labels[metrics.LabelNamespaceName.Key]
	if !found {
		return
	}

	nsObj, exists, err := this.store.GetByKey(namespaceName)
	if exists && err == nil {
		namespace, ok := nsObj.(*kube_api.Namespace)
		if ok {
			metricSet.Labels[metrics.LabelPodNamespaceUID.Key] = string(namespace.UID)
		} else {
			log.Errorf("Wrong namespace store content")
		}
	} else if err != nil {
		log.Warningf("Failed to get namespace %s: %v", namespaceName, err)
	} else if !exists {
		log.Warningf("Namespace doesn't exist: %s", namespaceName)
	}
}

func NewNamespaceBasedEnricher(url *url.URL) (*NamespaceBasedEnricher, error) {
	kubeConfig, err := kube_config.GetKubeClientConfig(url)
	if err != nil {
		return nil, err
	}
	kubeClient := kube_client.NewForConfigOrDie(kubeConfig)
	return &NamespaceBasedEnricher{
		store: util.GetNamespaceStore(kubeClient),
	}, nil
}
