// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package telegraf

import (
	"fmt"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"strings"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/telegraf"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultEncoder = telegrafEncoder{}

func NewTargetHandler(plugin string, handler metrics.ProviderHandler) discovery.TargetHandler {
	registryName := strings.Replace(plugin, "/", ".", -1)
	return discovery.NewHandler(
		discovery.ProviderInfo{
			Handler: handler,
			Factory: telegraf.NewFactory(),
			Encoder: defaultEncoder,
		},
		discovery.NewRegistry(registryName),
	)
}

type telegrafEncoder struct{}

func NewEncoder() discovery.Encoder {
	return telegrafEncoder{}
}

func (e telegrafEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) (interface{}, bool) {
	if ip == "" || ip == "None" {
		return configuration.TelegrafSourceConfig{}, false
	}

	result := configuration.TelegrafSourceConfig{
		Transforms: configuration.Transforms{
			Tags: make(map[string]string),
		},
	}

	if kind == discovery.ServiceType.String() {
		// always use leader election for cluster level resources
		result.UseLeaderElection = true
	}

	// panics if rule is not of expected type
	cfg := rule.(discovery.PluginConfig)
	name := discovery.ResourceName(kind, meta)
	pluginName := strings.Replace(cfg.Type, "telegraf/", "", -1)

	result.Discovered = "rule"
	result.Plugins = []string{pluginName}
	result.Name = name

	// parse telegraf configuration
	scheme := utils.Param(meta, "", cfg.Scheme, "http")
	server := fmt.Sprintf("%s://%s:%s", scheme, ip, cfg.Port)
	conf := strings.Replace(cfg.Conf, "${server}", server, -1)
	conf = strings.Replace(conf, "${host}", ip, -1)
	conf = strings.Replace(conf, "${port}", cfg.Port, -1)
	result.Conf = conf

	// parse prefix, tags, labels and filters
	prefix := utils.Param(meta, discovery.PrefixAnnotation, cfg.Prefix, "")
	includeLabels := utils.Param(meta, discovery.LabelsAnnotation, cfg.IncludeLabels, "true")

	result.Prefix = prefix
	result.Collection = configuration.CollectionConfig{
		Interval: cfg.Collection.Interval,
		Timeout:  cfg.Collection.Timeout,
	}

	utils.EncodeMeta(result.Tags, kind, meta)
	utils.EncodeTags(result.Tags, "", cfg.Tags)
	if includeLabels == "true" {
		utils.EncodeTags(result.Tags, "label.", meta.Labels)
	}
	result.Filters = cfg.Filters

	return result, true
}
