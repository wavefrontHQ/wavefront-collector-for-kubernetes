package prometheus

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	scrapeAnnotation = "prometheus.io/scrape"
	schemeAnnotation = "prometheus.io/scheme"
	pathAnnotation   = "prometheus.io/path"
	portAnnotation   = "prometheus.io/port"
	prefixAnnotation = "prometheus.io/prefix"
	labelsAnnotation = "prometheus.io/includeLabels"
	sourceAnnotation = "prometheus.io/source"
)

// used as source for discovered resources
var nodeName string

func init() {
	nodeName = util.GetNodeName()
}

type prometheusEncoder struct{}

func (e prometheusEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) url.Values {
	cfg := discovery.PrometheusConfig{}
	if rule != nil {
		cfg = rule.(discovery.PrometheusConfig)
	}
	return scrapeURL(ip, kind, meta, cfg)
}

func scrapeURL(ip, kind string, meta metav1.ObjectMeta, rule discovery.PrometheusConfig) url.Values {
	if ip == "" {
		glog.V(5).Infof("missing ip for %s=%s", kind, meta.Name)
		return url.Values{}
	}

	scrape := utils.Param(meta, scrapeAnnotation, "", "false")
	if rule.Name == "" && scrape != "true" {
		glog.V(5).Infof("scrape=false for %s=%s annotations=%q", kind, meta.Name, meta.Annotations)
		return url.Values{}
	}

	values := url.Values{}
	values.Set("discovered", "true")

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
	return values
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
		glog.V(5).Infof("using port 8080 for %s", name)
		return "8080"
	}
	return port
}
