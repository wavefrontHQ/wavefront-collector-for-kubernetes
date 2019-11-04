// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

// handles runtime changes to plugin rules
type ruleHandler struct {
	d                *discoverer
	ph               metrics.ProviderHandler
	daemon           bool
	kubeClient       kubernetes.Interface
	lister           discovery.ResourceLister
	annotationPrefix string
	rulesCount       gm.Gauge
}

// Gets a new rule handler that can handle runtime changes to plugin rules
func newRuleHandler(d discovery.Discoverer, cfg RunConfig) discovery.RuleHandler {
	rh := &ruleHandler{
		d:                d.(*discoverer),
		ph:               cfg.Handler,
		daemon:           cfg.Daemon,
		kubeClient:       cfg.KubeClient,
		lister:           cfg.Lister,
		annotationPrefix: cfg.DiscoveryConfig.AnnotationPrefix,
		rulesCount:       gm.GetOrRegisterGauge("discovery.rules.count", gm.DefaultRegistry),
	}
	count := int64(len(rh.d.delegates))
	rh.rulesCount.Update(count)
	return rh
}

func (rh *ruleHandler) HandleAll(plugins []discovery.PluginConfig) error {
	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()

	// delete rules that were removed/renamed
	rules := make(map[string]bool, len(plugins))
	for _, rule := range plugins {
		rules[rule.Name] = true
	}
	for name := range rh.d.delegates {
		if _, exists := rules[name]; !exists {
			log.WithField("name", name).Info("deleting discovery rule")
			rh.internalDelete(name)
		}
	}

	// process current rules
	for _, rule := range plugins {
		err := rh.internalHandle(rule)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"name":  rule.Name,
				"type":  rule.Type,
			}).Error("error processing rule")
		}
	}
	rh.rulesCount.Update(int64(len(plugins)))
	return nil
}

func (rh *ruleHandler) Handle(plugin discovery.PluginConfig) error {
	log.Infof("handling rule=%s type=%s", plugin.Name, plugin.Type)
	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()
	return rh.internalHandle(plugin)
}

func (rh *ruleHandler) DeleteAll() {
	err := rh.HandleAll([]discovery.PluginConfig{})
	if err != nil {
		log.Errorf("error deleting rules: %v", err)
	}
	for _, rh := range rh.d.runtimeHandlers {
		rh.DeleteMissing(map[string]bool{})
	}
}

func (rh *ruleHandler) Delete(name string) {
	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()
	rh.internalDelete(name)
}

func (rh *ruleHandler) Count() int {
	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()
	return len(rh.d.delegates)
}

// this method should only be invoked after acquiring the internal lock
func (rh *ruleHandler) internalHandle(plugin discovery.PluginConfig) error {
	delegate, exists := rh.d.delegates[plugin.Name]
	if !exists {
		var err error
		delegate, err = makeDelegate(rh.ph, plugin, rh.annotationPrefix)
		if err != nil {
			return err
		}
		rh.d.delegates[plugin.Name] = delegate
	} else {
		// replace the delegate plugin and filter without changing the handler
		filter, err := newResourceFilter(plugin)
		if err != nil {
			return err
		}
		delegate.filter = filter
		delegate.plugin = plugin
	}

	if plugin.Selectors.ResourceType == discovery.NodeType.String() {
		return rh.discoverNodeEndpoint(plugin, delegate.handler)
	}
	return nil
}

// this method should only be invoked after acquiring the internal lock
func (rh *ruleHandler) internalDelete(name string) {
	// deletes relevant discoverer delegate
	if delegate, exists := rh.d.delegates[name]; exists {
		delegate.handler.DeleteMissing(nil)
		delete(rh.d.delegates, name)
	}
}

func (rh *ruleHandler) discoverNodeEndpoint(plugin discovery.PluginConfig, handler discovery.TargetHandler) error {
	nodes, err := rh.lister.ListNodes()
	if err != nil {
		return fmt.Errorf("error listing nodes: %v", err)
	}

	count := len(nodes)
	if count != 1 {
		// node based discovery not supported in non-daemonset mode
		return fmt.Errorf("invalid number of nodes found: %d", count)
	}

	_, ip, err := util.GetNodeHostnameAndIP(nodes[0])
	if err != nil {
		return fmt.Errorf("error getting node IP: %v", err)
	}

	//TODO: probe if the node:port is reachable

	resource := discovery.Resource{
		Kind: discovery.NodeType.String(),
		IP:   ip.String(),
		Meta: metav1.ObjectMeta{Name: util.GetNodeName()},
	}
	handler.Handle(resource, plugin)

	return nil
}
