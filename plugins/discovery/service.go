// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type serviceHandler struct {
	ch       chan struct{}
	informer cache.SharedInformer
}

func newServiceHandler(kubeClient kubernetes.Interface, discoverer discovery.Discoverer) *serviceHandler {
	s := kubeClient.CoreV1().Services(v1.NamespaceAll)
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return s.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return s.Watch(options)
		},
	}
	inf := cache.NewSharedInformer(lw, &v1.Service{}, 1*time.Hour)

	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateServiceIfValid(obj, discoverer)
		},
		UpdateFunc: func(_, obj interface{}) {
			updateServiceIfValid(obj, discoverer)
		},
		DeleteFunc: func(obj interface{}) {
			deleteServiceIfValid(obj, discoverer)
		},
	})
	return &serviceHandler{
		informer: inf,
	}
}

func deleteServiceIfValid(obj interface{}, discoverer discovery.Discoverer) {
	service, ok := obj.(*v1.Service)
	if ok {
		discoverer.Discover(discovery.Resource{
			Kind: discovery.ServiceType.String(),
			IP:   service.Spec.ClusterIP,
			Meta: service.ObjectMeta,
		})
	}
}

func updateServiceIfValid(obj interface{}, discoverer discovery.Discoverer) {
	service, ok := obj.(*v1.Service)
	if ok && hasIP(service.Spec.ClusterIP) {
		discoverer.Discover(discovery.Resource{
			Kind: discovery.ServiceType.String(),
			IP:   service.Spec.ClusterIP,
			Meta: service.ObjectMeta,
		})
	}
}

func hasIP(ip string) bool {
	return ip != "" && ip != "None"
}

func (handler *serviceHandler) start() {
	handler.ch = make(chan struct{})
	go handler.informer.Run(handler.ch)
}

func (handler *serviceHandler) stop() {
	if handler.ch != nil {
		close(handler.ch)
	}
}
