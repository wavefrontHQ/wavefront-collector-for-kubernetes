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

package util

import (
	"github.com/golang/glog"
	"os"
	"sync"
	"time"

	kube_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	kube_client "k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	NodeNameEnvVar      = "POD_NODE_NAME"
	NamespaceNameEnvVar = "POD_NAMESPACE_NAME"
)

var (
	lock       sync.Mutex
	nodeLister v1listers.NodeLister
	reflector  *cache.Reflector
)

func GetNodeLister(kubeClient *kube_client.Clientset) (v1listers.NodeLister, *cache.Reflector, error) {
	lock.Lock()
	defer lock.Unlock()

	// init just one instance per collector agent
	if nodeLister != nil {
		return nodeLister, reflector, nil
	}

	fieldSelector := GetFieldSelector("nodes")
	lw := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "nodes", kube_api.NamespaceAll, fieldSelector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	nodeLister = v1listers.NewNodeLister(store)
	reflector = cache.NewReflector(lw, &kube_api.Node{}, store, time.Hour)
	go reflector.Run(wait.NeverStop)
	return nodeLister, reflector, nil
}

//TODO: verify how these listers are used
func GetPodLister(kubeClient *kube_client.Clientset) (v1listers.PodLister, error) {
	fieldSelector := GetFieldSelector("pods")
	lw := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", kube_api.NamespaceAll, fieldSelector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1listers.NewPodLister(store)
	reflector := cache.NewReflector(lw, &kube_api.Pod{}, store, time.Hour)
	go reflector.Run(wait.NeverStop)
	return podLister, nil
}

func GetServiceLister(kubeClient *kube_client.Clientset) (v1listers.ServiceLister, error) {
	lw := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "services", kube_api.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	serviceLister := v1listers.NewServiceLister(store)
	reflector := cache.NewReflector(lw, &kube_api.Service{}, store, time.Hour)
	go reflector.Run(wait.NeverStop)
	return serviceLister, nil
}

func GetFieldSelector(resourceType string) fields.Selector {
	fieldSelector := fields.Everything()
	nodeName := GetNodeName()
	if nodeName != "" {
		switch resourceType {
		case "pods":
			fieldSelector = fields.ParseSelectorOrDie("spec.nodeName=" + nodeName)
		case "nodes":
			fieldSelector = fields.OneTermEqualSelector("metadata.name", nodeName)
		default:
			glog.Infof("invalid resource type: %s", resourceType)
		}
	}
	glog.V(2).Infof("using fieldSelector: %q for resourceType: %s", fieldSelector, resourceType)
	return fieldSelector
}

func GetNodeName() string {
	return os.Getenv(NodeNameEnvVar)
}

func GetNamespaceName() string {
	return os.Getenv(NamespaceNameEnvVar)
}
