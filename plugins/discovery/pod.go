// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type podHandler struct {
	ch       chan struct{}
	informer cache.SharedInformer
}

func newPodHandler(kubeClient kubernetes.Interface, discoverer discovery.Discoverer) *podHandler {
	client := kubeClient.CoreV1().RESTClient()
	fieldSelector := util.GetFieldSelector("pods")
	lw := cache.NewListWatchFromClient(client, "pods", v1.NamespaceAll, fieldSelector)
	inf := cache.NewSharedInformer(lw, &v1.Pod{}, 1*time.Hour)

	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updatePodIfValid(obj, discoverer)
		},
		UpdateFunc: func(_, obj interface{}) {
			updatePodIfValid(obj, discoverer)
		},
		DeleteFunc: func(obj interface{}) {
			deletePodIfValid(obj, discoverer)
		},
	})
	return &podHandler{
		informer: inf,
	}
}

func deletePodIfValid(obj interface{}, discoverer discovery.Discoverer) {
	pod, ok := obj.(*v1.Pod)
	if ok {
		discoverer.Delete(discovery.Resource{
			Kind:       discovery.PodType.String(),
			IP:         pod.Status.PodIP,
			Meta:       pod.ObjectMeta,
			Containers: pod.Spec.Containers,
		})
	}
}

func updatePodIfValid(obj interface{}, discoverer discovery.Discoverer) {
	pod, ok := obj.(*v1.Pod)
	if ok { podUpdated( pod, discoverer ) }
}

func podUpdated(pod *v1.Pod, discoverer discovery.Discoverer) {
	if podReady(pod) {
		discoverer.Discover(discovery.Resource{
			Kind:       discovery.PodType.String(),
			IP:         pod.Status.PodIP,
			Meta:       pod.ObjectMeta,
			Containers: pod.Spec.Containers,
		})
	}
}

func podReady(pod *v1.Pod) bool {
	if pod.Status.Phase != "Running" {
		return false
	}
	if pod.Status.PodIP == "" || pod.Status.PodIP == "None" {
		return false
	}
	return true
}

func (handler *podHandler) start() {
	handler.ch = make(chan struct{})
	go handler.informer.Run(handler.ch)
}

func (handler *podHandler) stop() {
	if handler.ch != nil {
		close(handler.ch)
	}
}
