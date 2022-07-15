// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	kube_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	NodeNameEnvVar           = "POD_NODE_NAME"
	NamespaceNameEnvVar      = "POD_NAMESPACE_NAME"
	InstallationMethodEnvVar = "INSTALLATION_METHOD"
	ForceGC                  = "FORCE_GC"
	KubernetesVersionEnvVar  = "KUBERNETES_VERSION"
	KubernetesProviderEnvVar = "KUBERNETES_PROVIDER"
)

const (
	POD_PHASE_PENDING = iota + 1
	POD_PHASE_RUNNING
	POD_PHASE_SUCCEEDED
	POD_PHASE_FAILED
	POD_PHASE_UNKNOWN
)

const (
	CONTAINER_STATE_RUNNING = iota + 1
	CONTAINER_STATE_WAITING
	CONTAINER_STATE_TERMINATED
)

var (
	lock       sync.Mutex
	nodeLister v1listers.NodeLister
	reflector  *cache.Reflector
	podLister  v1listers.PodLister
	nsStore    cache.Store
	agentType  AgentType
)

type AgentType interface {
	ScrapeCluster() bool
	ScrapeAnyNodes() bool
	ScrapeOnlyOwnNode() bool
}

func GetNodeLister(kubeClient kubernetes.Interface) (v1listers.NodeLister, *cache.Reflector, error) {
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
	go reflector.Run(NeverStop)
	return nodeLister, reflector, nil
}

func GetPodLister(kubeClient kubernetes.Interface) (v1listers.PodLister, error) {
	lock.Lock()
	defer lock.Unlock()

	// init just one instance per collector agent
	if podLister != nil {
		return podLister, nil
	}

	fieldSelector := GetFieldSelector("pods")
	lw := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", kube_api.NamespaceAll, fieldSelector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister = v1listers.NewPodLister(store)
	reflector := cache.NewReflector(lw, &kube_api.Pod{}, store, time.Hour)
	go reflector.Run(NeverStop)
	return podLister, nil
}

func GetServiceLister(kubeClient kubernetes.Interface) (v1listers.ServiceLister, error) {
	lw := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "services", kube_api.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	serviceLister := v1listers.NewServiceLister(store)
	reflector := cache.NewReflector(lw, &kube_api.Service{}, store, time.Hour)
	go reflector.Run(NeverStop)
	return serviceLister, nil
}

func GetNamespaceStore(kubeClient kubernetes.Interface) cache.Store {
	lock.Lock()
	defer lock.Unlock()

	// init just once per collector agent
	if nsStore != nil {
		return nsStore
	}

	lw := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "namespaces", kube_api.NamespaceAll, fields.Everything())
	nsStore = cache.NewStore(cache.MetaNamespaceKeyFunc)
	reflector := cache.NewReflector(lw, &kube_api.Namespace{}, nsStore, time.Hour)
	go reflector.Run(NeverStop)
	return nsStore
}

func GetFieldSelector(resourceType string) fields.Selector {
	fieldSelector := fields.Everything()
	nodeName := GetNodeName()
	if ScrapeOnlyOwnNode() && nodeName != "" {
		switch resourceType {
		case "pods":
			fieldSelector = fields.ParseSelectorOrDie("spec.nodeName=" + nodeName)
		case "nodes":
			fieldSelector = fields.OneTermEqualSelector("metadata.name", nodeName)
		default:
			log.Infof("invalid resource type: %s", resourceType)
		}
	}
	log.Debugf("using fieldSelector: %q for resourceType: %s", fieldSelector, resourceType)
	return fieldSelector
}

func ScrapeCluster() bool {
	return agentType.ScrapeCluster()
}

func SetAgentType(value AgentType) {
	agentType = value
}

func ScrapeAnyNodes() bool {
	return agentType.ScrapeAnyNodes()
}

func ScrapeOnlyOwnNode() bool {
	return agentType.ScrapeOnlyOwnNode()
}

func GetNodeName() string {
	return os.Getenv(NodeNameEnvVar)
}

func GetNamespaceName() string {
	return os.Getenv(NamespaceNameEnvVar)
}

func GetInstallationMethod() string {

	installationMethod := os.Getenv(InstallationMethodEnvVar)
	if len(installationMethod) == 0 {
		return "unknown"
	}
	return installationMethod
}

func GetKubernetesProvider() string {
	return os.Getenv(KubernetesProviderEnvVar)
}

func SetKubernetesVersion(version string) {
	os.Setenv(KubernetesVersionEnvVar, version)
}

func GetKubernetesVersion() string {
	return os.Getenv(KubernetesVersionEnvVar)
}

func SetKubernetesProvider(providerID string) {
	provider := strings.Split(providerID, ":")
	if len(provider[0]) > 0 {
		os.Setenv(KubernetesProviderEnvVar, provider[0])
	} else {
		os.Setenv(KubernetesProviderEnvVar, "unknown")
	}
}

func AddK8sTags(tags map[string]string) {
	// Use separate function to add K8s tags since the Env variables are set via summary source
	if len(tags["k8s_version"]) == 0 && len(GetKubernetesVersion()) > 0 {
		tags["k8s_version"] = GetKubernetesVersion()
	}
	if len(tags["k8s_provider"]) == 0 && len(GetKubernetesProvider()) > 0 {
		tags["k8s_provider"] = GetKubernetesProvider()
	}
}

func GetNodeHostnameAndIP(node *kube_api.Node) (string, net.IP, error) {
	for _, c := range node.Status.Conditions {
		if c.Type == kube_api.NodeReady && c.Status != kube_api.ConditionTrue {
			return "", nil, fmt.Errorf("node %v is not ready", node.Name)
		}
	}
	hostname, ip := node.Name, ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == kube_api.NodeHostName && addr.Address != "" {
			hostname = addr.Address
		}
		if addr.Type == kube_api.NodeInternalIP && addr.Address != "" {
			if net.ParseIP(addr.Address) != nil {
				ip = addr.Address
			}
		}
		if addr.Type == kube_api.NodeExternalIP && addr.Address != "" && ip == "" {
			ip = addr.Address
		}
	}
	if parsedIP := net.ParseIP(ip); parsedIP != nil {
		return hostname, parsedIP, nil
	}
	return "", nil, fmt.Errorf("node %v has no valid hostname and/or IP address: %v %v", node.Name, hostname, ip)
}

func GetNodeRole(node *kube_api.Node) string {
	if _, ok := node.GetLabels()["node-role.kubernetes.io/control-plane"]; ok {
		return "control-plane"
	}
	if _, ok := node.GetLabels()["node-role.kubernetes.io/master"]; ok {
		return "control-plane"
	}
	return "worker"
}

type ContainerStateInfo struct {
	Value    int
	State    string
	Reason   string
	ExitCode int32
}

func (csi ContainerStateInfo) IsKnownState() bool {
	return csi.Value > 0
}

func (csi ContainerStateInfo) AddMetricTags(tags map[string]string) {
	if csi.IsKnownState() {
		tags["status"] = csi.State
		if csi.Reason != "" {
			tags["reason"] = csi.Reason
			tags["exit_code"] = fmt.Sprint(csi.ExitCode)
		}
	}
}

func NewContainerStateInfo(state kube_api.ContainerState) ContainerStateInfo {
	if state.Running != nil {
		return ContainerStateInfo{
			Value:    CONTAINER_STATE_RUNNING,
			State:    "running",
			Reason:   "",
			ExitCode: 0,
		}
	}
	if state.Waiting != nil {
		return ContainerStateInfo{
			Value:    CONTAINER_STATE_WAITING,
			State:    "waiting",
			Reason:   state.Waiting.Reason,
			ExitCode: 0,
		}
	}
	if state.Terminated != nil {
		return ContainerStateInfo{
			Value:    CONTAINER_STATE_TERMINATED,
			State:    "terminated",
			Reason:   state.Terminated.Reason,
			ExitCode: state.Terminated.ExitCode,
		}
	}
	return ContainerStateInfo{
		Value:    0,
		State:    "unknown",
		Reason:   "",
		ExitCode: 0,
	}
}

func ConvertPodPhase(phase kube_api.PodPhase) int64 {
	switch phase {
	case kube_api.PodPending:
		return POD_PHASE_PENDING
	case kube_api.PodRunning:
		return POD_PHASE_RUNNING
	case kube_api.PodSucceeded:
		return POD_PHASE_SUCCEEDED
	case kube_api.PodFailed:
		return POD_PHASE_FAILED
	case kube_api.PodUnknown:
		return POD_PHASE_UNKNOWN
	default:
		return POD_PHASE_UNKNOWN
	}
}
