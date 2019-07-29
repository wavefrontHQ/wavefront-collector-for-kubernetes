package prometheus

import (
	"fmt"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/httputil"
	"net/url"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
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

func (e prometheusEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, cfg interface{}) url.Values {
	values := url.Values{}
	rule := discovery.PluginConfig{}
	discoveryType := "annotation"
	if cfg != nil {
		rule = cfg.(discovery.PluginConfig)
		discoveryType = "rule"
		collectionInterval := utils.Param(meta, collectionIntervalAnnotation, rule.Collection.Interval.String(), "0s")
		values.Set("collectionInterval", collectionInterval)
		timeout := utils.Param(meta, timeoutAnnotation, rule.Collection.Timeout.String(), "0s")
		values.Set("timeout", timeout)
	}
	values.Set("discovered", discoveryType)

	if ip == "" {
		log.Debugf("missing ip for %s=%s", kind, meta.Name)
		return url.Values{}
	}

	scrape := utils.Param(meta, scrapeAnnotation, "", "false")
	if rule.Name == "" && scrape != "true" {
		log.Debugf("prometheus scrape=false for %s=%s", kind, meta.Name)
		return url.Values{}
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

	encodeBase(values, scheme, ip, port, path, name, source, prefix)
	utils.EncodeMeta(values, kind, meta)
	utils.EncodeTags(values, "", rule.Tags)
	if includeLabels == "true" {
		utils.EncodeTags(values, "label.", meta.Labels)
	}
	utils.EncodeFilters(values, rule.Filters)

	err := encodeConf(values, rule.Conf)
	if err != nil {
		return url.Values{}
	}
	return values
}

func encodeConf(values url.Values, conf string) error {
	if conf != "" {
		httpConf, err := httputil.FromYAML([]byte(conf))
		if err != nil {
			return err
		}
		utils.EncodeHTTPConfig(values, httpConf)
	}
	return nil
}

func encodeBase(values url.Values, scheme, ip, port, path, name, source, prefix string) {
	if port != "" {
		port = fmt.Sprintf(":%s", port)
	}
	values.Set("url", fmt.Sprintf("%s://%s%s%s", scheme, ip, port, path))
	values.Add("name", name)

	if source != "" {
		values.Add("source", source)
	}
	if prefix != "" {
		values.Add("prefix", prefix)
	}
}

func sanitizePort(name, port string) string {
	if strings.Contains(name, "kube-state-metrics") && port == "" {
		log.Debugf("using port 8080 for %s", name)
		return "8080"
	}
	return port
}
