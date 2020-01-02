// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	DaemonSets               = "daemonsets"
	Deployments              = "deployments"
	ReplicaSets              = "replicasets"
	StatefulSets             = "statefulsets"
	Jobs                     = "jobs"
	CronJobs                 = "cronjobs"
	HorizontalPodAutoscalers = "horizontalpodautoscalers"
)

var (
	doOnce    sync.Once
	singleton *Lister
)

type Lister struct {
	kubeClient kubernetes.Interface
	stores     map[string]cache.Store
}

func NewLister(kubeClient kubernetes.Interface) *Lister {
	doOnce.Do(func() {
		singleton = &Lister{
			kubeClient: kubeClient,
			stores:     buildStores(kubeClient),
		}
	})
	return singleton
}

func buildStores(kubeClient kubernetes.Interface) map[string]cache.Store {
	m := make(map[string]cache.Store)
	m[DaemonSets] = buildStore(DaemonSets, &appsv1.DaemonSet{}, kubeClient.AppsV1().RESTClient())
	m[Deployments] = buildStore(Deployments, &appsv1.Deployment{}, kubeClient.AppsV1().RESTClient())
	m[StatefulSets] = buildStore(StatefulSets, &appsv1.StatefulSet{}, kubeClient.AppsV1().RESTClient())
	m[ReplicaSets] = buildStore(ReplicaSets, &appsv1.ReplicaSet{}, kubeClient.AppsV1().RESTClient())
	m[Jobs] = buildStore(Jobs, &batchv1.Job{}, kubeClient.BatchV1().RESTClient())
	m[CronJobs] = buildStore(CronJobs, &batchv1beta1.CronJob{}, kubeClient.BatchV1beta1().RESTClient())
	m[HorizontalPodAutoscalers] = buildStore(HorizontalPodAutoscalers, &v2beta1.HorizontalPodAutoscaler{}, kubeClient.AutoscalingV2beta1().RESTClient())
	return m
}

func buildStore(resource string, resType runtime.Object, getter cache.Getter) cache.Store {
	lw := cache.NewListWatchFromClient(getter, resource, v1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	reflector := cache.NewReflector(lw, resType, store, time.Hour)
	go reflector.Run(wait.NeverStop)
	return store
}

func (l *Lister) List(resource string) ([]interface{}, error) {
	if store, exists := l.stores[resource]; exists {
		return store.List(), nil
	} else {
		return nil, fmt.Errorf("unsupported resource type: %s", resource)
	}
}
