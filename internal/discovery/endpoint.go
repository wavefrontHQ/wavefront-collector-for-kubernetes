// Copyright 2019-2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	gm "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
)

type ProviderInfo struct {
	Handler metrics.ProviderHandler
	Factory metrics.ProviderFactory
	Encoder Encoder
}

type defaultEndpointHandler struct {
	providers map[string]ProviderInfo
	counts    map[string]gm.Counter
}

func NewEndpointHandler(providers map[string]ProviderInfo) EndpointHandler {
	counts := make(map[string]gm.Counter, len(providers))
	for k := range providers {
		key := reporting.EncodeKey("discovery.targets.registered", map[string]string{"type": k})
		counts[k] = gm.GetOrRegisterCounter(key, gm.DefaultRegistry)
	}
	return &defaultEndpointHandler{
		providers: providers,
		counts:    counts,
	}
}

func (d *defaultEndpointHandler) Add(ep *Endpoint) {
	if delegate, ok := d.providers[ep.PluginType]; ok {
		provider, err := delegate.Factory.Build(ep.Config)
		if err != nil {
			log.Error(err)
			return
		}
		delegate.Handler.AddProvider(provider)
		d.counts[ep.PluginType].Inc(1)
	} else {
		log.WithFields(log.Fields{
			"endpoint": ep.Name,
			"type":     ep.PluginType,
		}).Error("failed to add endpoint")
	}
}

func (d *defaultEndpointHandler) Delete(ep *Endpoint) {
	log.WithFields(log.Fields{
		"endpoint": ep.Name,
		"type":     ep.PluginType,
	}).Debug("deleting endpoint")

	if delegate, ok := d.providers[ep.PluginType]; ok {
		name := fmt.Sprintf("%s: %s", delegate.Factory.Name(), ep.Name)
		delegate.Handler.DeleteProvider(name)
		d.counts[ep.PluginType].Dec(1)
	} else {
		log.WithFields(log.Fields{
			"endpoint": ep.Name,
			"type":     ep.PluginType,
		}).Error("failed to delete endpoint")
	}
}

func pluginType(plugin PluginConfig) string {
	if strings.Contains(plugin.Type, "prometheus") {
		return "prometheus"
	} else if strings.Contains(plugin.Type, "telegraf") {
		return "telegraf"
	}
	return ""
}
