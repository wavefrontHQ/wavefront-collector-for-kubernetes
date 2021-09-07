// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	scrapeAnnotationFormat             = "%s/scrape"
	schemeAnnotationFormat             = "%s/scheme"
	pathAnnotationFormat               = "%s/path"
	portAnnotationFormat               = "%s/port"
	prefixAnnotationFormat             = "%s/prefix"
	labelsAnnotationFormat             = "%s/includeLabels"
	sourceAnnotationFormat             = "%s/source"
	collectionIntervalAnnotationFormat = "%s/collectionInterval"
	timeoutAnnotationFormat            = "%s/timeout"
	insecureSkipVerifyFormat           = "%s/insecureSkipVerify"
	serverNameFormat                   = "%s/serverName"
)

// used as source for discovered resources
var nodeName string

func init() {
	nodeName = util.GetNodeName()
}

type prometheusEncoder struct {
	scrapeAnnotation             string
	schemeAnnotation             string
	pathAnnotation               string
	portAnnotation               string
	prefixAnnotation             string
	labelsAnnotation             string
	sourceAnnotation             string
	collectionIntervalAnnotation string
	timeoutAnnotation            string
	insecureSkipVerifyAnnotation string
	serverNameAnnotation         string
}

func newPrometheusEncoder(prefix string) prometheusEncoder {
	if len(prefix) == 0 {
		prefix = "prometheus.io"
	}
	return prometheusEncoder{
		scrapeAnnotation:             customAnnotation(scrapeAnnotationFormat, prefix),
		schemeAnnotation:             customAnnotation(schemeAnnotationFormat, prefix),
		pathAnnotation:               customAnnotation(pathAnnotationFormat, prefix),
		portAnnotation:               customAnnotation(portAnnotationFormat, prefix),
		prefixAnnotation:             customAnnotation(prefixAnnotationFormat, prefix),
		labelsAnnotation:             customAnnotation(labelsAnnotationFormat, prefix),
		sourceAnnotation:             customAnnotation(sourceAnnotationFormat, prefix),
		collectionIntervalAnnotation: customAnnotation(collectionIntervalAnnotationFormat, prefix),
		timeoutAnnotation:            customAnnotation(timeoutAnnotationFormat, prefix),
		insecureSkipVerifyAnnotation: customAnnotation(insecureSkipVerifyFormat, prefix),
		serverNameAnnotation:         customAnnotation(serverNameFormat, prefix),
	}
}

func (e prometheusEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, cfg interface{}) (string, interface{}, bool) {
	if ip == "" || ip == "None" {
		log.Debugf("missing ip for %s=%s", kind, meta.Name)
		return "", configuration.PrometheusSourceConfig{}, false
	}

	result := configuration.PrometheusSourceConfig{
		Transforms: configuration.Transforms{
			Tags: make(map[string]string),
		},
	}

	rule := discovery.PluginConfig{}
	discoveryType := "annotation"
	if cfg != nil {
		rule = cfg.(discovery.PluginConfig)
		if rule.Name != "" {
			discoveryType = "rule"
		}
	}
	result.Discovered = discoveryType

	if kind != discovery.ServiceType.String() {
		result.PerNode = true
	}

	collectionInterval := utils.Param(meta, e.collectionIntervalAnnotation, rule.Collection.Interval.String(), "0s")
	timeout := utils.Param(meta, e.timeoutAnnotation, rule.Collection.Timeout.String(), "0s")

	collectionDuration, err := time.ParseDuration(collectionInterval)
	if err != nil {
		log.Errorf("error parsing collection interval: %s %v", collectionInterval, err)
		return "", result, false
	}
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		log.Errorf("error parsing timeout: %s %v", timeout, err)
		return "", result, false
	}
	result.Collection = configuration.CollectionConfig{
		Interval: collectionDuration,
		Timeout:  timeoutDuration,
	}

	scrape := utils.Param(meta, e.scrapeAnnotation, "", "false")
	if rule.Name == "" && scrape != "true" {
		log.Debugf("prometheus scrape=false for %s=%s", kind, meta.Name)
		return "", result, false
	}

	scheme := utils.Param(meta, e.schemeAnnotation, rule.Scheme, "http")
	path := utils.Param(meta, e.pathAnnotation, rule.Path, "/metrics")
	port := utils.Param(meta, e.portAnnotation, rule.Port, "")
	prefix := utils.Param(meta, e.prefixAnnotation, rule.Prefix, "")
	source := utils.Param(meta, e.sourceAnnotation, rule.Source, nodeName)
	includeLabels := utils.Param(meta, e.labelsAnnotation, rule.IncludeLabels, "true")
	insecureSkipVerify := utils.Param(meta, e.insecureSkipVerifyAnnotation, "", "")
	serverName := utils.Param(meta, e.serverNameAnnotation, "", "")

	if source == "" {
		source = meta.Name
	}
	name := discovery.ResourceName(kind, meta)
	port = sanitizePort(meta.Name, port)
	name = uniqueName(name, port, path)

	encodeBase(&result, scheme, ip, port, path, name, source, prefix)
	utils.EncodeMeta(result.Tags, kind, meta)
	utils.EncodeTags(result.Tags, "", rule.Tags)
	if includeLabels == "true" {
		utils.EncodeTags(result.Tags, "label.", meta.Labels)
	}
	result.Filters = rule.Filters

	err = encodeHTTPConf(&result, rule.Conf, insecureSkipVerify, serverName)
	if err != nil {
		return "", result, false
	}
	return name, result, true
}

func encodeHTTPConf(cfg *configuration.PrometheusSourceConfig, conf, insecure, serverName string) error {
	if conf != "" {
		httpConf, err := httputil.FromYAML([]byte(conf))
		if err != nil {
			return err
		}
		cfg.HTTPClientConfig = httpConf
	} else {
		insecureBool := true
		if len(insecure) > 0 {
			insecureBool, _ = strconv.ParseBool(insecure)
		} else if len(serverName) > 0 {
			insecureBool = false
		}
		cfg.HTTPClientConfig = httputil.ClientConfig{
			TLSConfig: httputil.TLSConfig{
				InsecureSkipVerify: insecureBool,
				ServerName:         serverName,
			},
		}
	}
	return nil
}

func encodeBase(cfg *configuration.PrometheusSourceConfig, scheme, ip, port, path, name, source, prefix string) {
	if port != "" {
		port = fmt.Sprintf(":%s", port)
	}
	cfg.URL = fmt.Sprintf("%s://%s%s%s", scheme, ip, port, path)
	cfg.Name = name
	cfg.Source = source
	cfg.Prefix = prefix
}

func sanitizePort(name, port string) string {
	if strings.Contains(name, "kube-state-metrics") && port == "" {
		log.Debugf("using port 8080 for %s", name)
		return "8080"
	}
	return port
}

func customAnnotation(annotationFormat, prefix string) string {
	return fmt.Sprintf(annotationFormat, prefix)
}

func uniqueName(name, port, path string) string {
	out := name
	if port != "" {
		out = fmt.Sprintf("%s:%s", out, port)
	}
	if path != "" {
		out = fmt.Sprintf("%s%s", out, path)
	}
	return out
}
