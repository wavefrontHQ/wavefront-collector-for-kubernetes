// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/discovery/prometheus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/discovery/telegraf"

	gm "github.com/rcrowley/go-metrics"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type delegate struct {
	filter *resourceFilter
	plugin discovery.PluginConfig
}

type discoverer struct {
	wg    sync.WaitGroup
	mtx   sync.RWMutex
	queue chan discovery.Resource

	lister    discovery.ResourceLister
	ruleCount gm.Gauge

	endpoints       map[string][]*discovery.Endpoint
	endpointHandler discovery.EndpointHandler

	endpointCreator endpointCreator
}

func newDiscoverer(handler metrics.ProviderHandler, discoveryCfg discovery.Config, lister discovery.ResourceLister) discovery.Discoverer {
	ec := endpointCreator{
		delegates:                  makeDelegates(discoveryCfg),
		annotationExcludes:         makeAnnotationExclusions(discoveryCfg.AnnotationExcludes),
		providers:                  makeProviders(handler, discoveryCfg),
		disableAnnotationDiscovery: discoveryCfg.DisableAnnotationDiscovery,
	}
	d := &discoverer{
		queue:           make(chan discovery.Resource, 1000),
		lister:          lister,
		ruleCount:       gm.GetOrRegisterGauge("discovery.rules.count", gm.DefaultRegistry),
		endpoints:       make(map[string][]*discovery.Endpoint, 32),
		endpointHandler: discovery.NewEndpointHandler(makeProviders(handler, discoveryCfg)),
		endpointCreator: ec,
	}
	d.ruleCount.Update(int64(len(d.endpointCreator.delegates)))
	go d.dequeue()
	go d.discoverNodeEndpoints(discoveryCfg.PluginConfigs)
	return d
}

func makeProviders(handler metrics.ProviderHandler, discoveryCfg discovery.Config) map[string]discovery.ProviderInfo {
	providers := make(map[string]discovery.ProviderInfo, 2)
	providers["prometheus"] = prometheus.NewProviderInfo(handler, discoveryCfg.AnnotationPrefix)
	providers["telegraf"] = telegraf.NewProviderInfo(handler)
	return providers
}

func makeDelegates(discoveryCfg discovery.Config) map[string]*delegate {
	plugins := discoveryCfg.PluginConfigs
	delegates := make(map[string]*delegate, len(plugins))
	for _, plugin := range plugins {
		delegate, err := makeDelegate(plugin)
		if err != nil {
			log.Errorf("error parsing plugin: %s error: %v", plugin.Name, err)
			continue
		}
		delegates[plugin.Name] = delegate
	}
	return delegates
}

func makeDelegate(plugin discovery.PluginConfig) (*delegate, error) {
	filter, err := newResourceFilter(plugin.Selectors)
	if err != nil {
		return nil, err
	}
	if plugin.Port != "" {
		_, err := strconv.ParseInt(plugin.Port, 10, 32)
		if err != nil {
			return nil, err
		}
	}
	if !(strings.Contains(plugin.Type, "prometheus") || strings.Contains(plugin.Type, "telegraf")) {
		return nil, fmt.Errorf("invalid plugin type: %s", plugin.Type)
	}
	return &delegate{
		filter: filter,
		plugin: plugin,
	}, nil
}

func makeAnnotationExclusions(selectors []discovery.Selectors) []*resourceFilter {
	var filters []*resourceFilter
	for _, selector := range selectors {
		filter, err := newResourceFilter(selector)
		if err != nil {
			log.Errorf("invalid annotation exclusion: %s", err.Error())
			continue
		}
		filters = append(filters, filter)
	}
	return filters
}

func (d *discoverer) enqueue(resource discovery.Resource) {
	d.wg.Add(1)
	defer d.wg.Done()
	d.queue <- resource
}

func (d *discoverer) dequeue() {
	for resource := range d.queue {
		switch resource.Status {
		case "delete":
			d.internalDelete(resource)
		default:
			d.internalDiscover(resource)
		}
	}
	log.Infof("stopping discoverer deque")
}

func (d *discoverer) Stop() {
	d.wg.Wait()
	close(d.queue)
}

func (d *discoverer) Discover(resource discovery.Resource) {
	d.enqueue(resource)
}

func (d *discoverer) Delete(resource discovery.Resource) {
	resource.Status = "delete"
	d.enqueue(resource)
}

func (d *discoverer) DeleteAll() {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	for k, eps := range d.endpoints {
		for _, ep := range eps {
			d.endpointHandler.Delete(ep)
		}
		delete(d.endpoints, k)
	}
}

func (d *discoverer) internalDiscover(resource discovery.Resource) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	eps := d.endpointCreator.discoverEndpoints(resource)

	resourceName := discovery.ResourceName(resource.Kind, resource.Meta)
	oldEps := d.endpoints[resourceName]
	delete(d.endpoints, resourceName)

	if len(eps) > 0 {
		d.endpoints[resourceName] = eps
	}

	if reflect.DeepEqual(eps, oldEps) {
		log.Debugf("no endpoint changes for %s", resourceName)
		return
	}

	for _, ep := range oldEps {
		d.endpointHandler.Delete(ep)
	}
	for _, ep := range eps {
		d.endpointHandler.Add(ep)
	}
}

func (d *discoverer) internalDelete(resource discovery.Resource) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	resourceName := discovery.ResourceName(resource.Kind, resource.Meta)
	eps := d.endpoints[resourceName]
	delete(d.endpoints, resourceName)

	for _, ep := range eps {
		d.endpointHandler.Delete(ep)
	}
}

func (d *discoverer) discoverNodeEndpoints(plugins []discovery.PluginConfig) {
	// wait for listers to index
	time.Sleep(30 * time.Second)

	for _, plugin := range plugins {
		if plugin.Selectors.ResourceType == discovery.NodeType.String() {
			err := d.discoverNodeEndpoint(plugin)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"name":  plugin.Name,
					"type":  plugin.Type,
				}).Error("error processing rule")
			}
		}
	}
}

func (d *discoverer) discoverNodeEndpoint(plugin discovery.PluginConfig) error {
	nodes, err := d.lister.ListNodes()
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

	resource := discovery.Resource{
		Kind: discovery.NodeType.String(),
		IP:   ip.String(),
		Meta: metav1.ObjectMeta{Name: util.GetNodeName()},
	}

	d.mtx.Lock()
	defer d.mtx.Unlock()

	if ep := d.endpointCreator.makeEndpoint(resource, plugin); ep != nil {
		name := discovery.ResourceName(resource.Kind, resource.Meta)
		var eps []*discovery.Endpoint
		if val, ok := d.endpoints[name]; ok {
			eps = val
		}
		eps = append(eps, ep)
		d.endpoints[name] = eps
		d.endpointHandler.Add(ep)
	}
	return nil
}

func pluginType(plugin discovery.PluginConfig) string {
	if strings.Contains(plugin.Type, "prometheus") {
		return "prometheus"
	} else if strings.Contains(plugin.Type, "telegraf") {
		return "telegraf"
	}
	return ""
}
