package prometheus

import (
	"fmt"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/httputil"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	scrapeAnnotation             = "prometheus.io/scrape"
	schemeAnnotation             = "prometheus.io/scheme"
	pathAnnotation               = "prometheus.io/path"
	portAnnotation               = "prometheus.io/port"
	prefixAnnotation             = "prometheus.io/prefix"
	labelsAnnotation             = "prometheus.io/includeLabels"
	sourceAnnotation             = "prometheus.io/source"
	collectionIntervalAnnotation = "prometheus.io/collectionInterval"
	timeoutAnnotation            = "prometheus.io/timeout"
)

// used as source for discovered resources
var nodeName string

func init() {
	nodeName = util.GetNodeName()
}

type prometheusEncoder struct{}

func (e prometheusEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, cfg interface{}) (interface{}, bool) {
	if ip == "" {
		log.Debugf("missing ip for %s=%s", kind, meta.Name)
		return configuration.PrometheusSourceConfig{}, false
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
		discoveryType = "rule"
		collectionInterval := utils.Param(meta, collectionIntervalAnnotation, rule.Collection.Interval.String(), "0s")
		timeout := utils.Param(meta, timeoutAnnotation, rule.Collection.Timeout.String(), "0s")

		collectionDuration, err := time.ParseDuration(collectionInterval)
		if err != nil {
			log.Errorf("error parsing collection interval: %s %v", collectionInterval, err)
			return result, false
		}
		timeoutDuration, err := time.ParseDuration(timeout)
		if err != nil {
			log.Errorf("error parsing timeout: %s %v", timeout, err)
			return result, false
		}
		result.Collection = configuration.CollectionConfig{
			Interval: collectionDuration,
			Timeout:  timeoutDuration,
		}
	}
	result.Discovered = discoveryType

	scrape := utils.Param(meta, scrapeAnnotation, "", "false")
	if rule.Name == "" && scrape != "true" {
		log.Debugf("prometheus scrape=false for %s=%s", kind, meta.Name)
		return result, false
	}

	scheme := utils.Param(meta, schemeAnnotation, rule.Scheme, "http")
	path := utils.Param(meta, pathAnnotation, rule.Path, "/metrics")
	port := utils.Param(meta, portAnnotation, rule.Port, "")
	prefix := utils.Param(meta, prefixAnnotation, rule.Prefix, "")
	source := utils.Param(meta, sourceAnnotation, rule.Source, nodeName)
	includeLabels := utils.Param(meta, labelsAnnotation, rule.IncludeLabels, "true")

	if source == "" {
		source = meta.Name
	}
	name := discovery.ResourceName(kind, meta)
	port = sanitizePort(meta.Name, port)

	encodeBase(&result, scheme, ip, port, path, name, source, prefix)
	utils.EncodeMeta(result.Tags, kind, meta)
	utils.EncodeTags(result.Tags, "", rule.Tags)
	if includeLabels == "true" {
		utils.EncodeTags(result.Tags, "label.", meta.Labels)
	}
	result.Filters = rule.Filters

	err := encodeConf(&result, rule.Conf)
	if err != nil {
		return result, false
	}
	return result, true
}

func encodeConf(cfg *configuration.PrometheusSourceConfig, conf string) error {
	if conf != "" {
		httpConf, err := httputil.FromYAML([]byte(conf))
		if err != nil {
			return err
		}
		cfg.HTTPClientConfig = httpConf
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
