package prometheus

import (
	"fmt"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
	"sort"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"

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

func scrapeURL(ip, kind string, meta metav1.ObjectMeta, rule discovery.PrometheusConfig) string {
	if ip == "" {
		glog.V(5).Infof("missing ip for %s=%s", kind, meta.Name)
		return ""
	}
	scrape := param(meta, scrapeAnnotation, "", "false")
	if rule.Name == "" && scrape != "true" {
		glog.V(5).Infof("scrape=false for %s=%s annotations=%q", kind, meta.Name, meta.Annotations)
		return ""
	}

	scheme := param(meta, schemeAnnotation, rule.Scheme, "http")
	path := param(meta, pathAnnotation, rule.Path, "/metrics")
	port := param(meta, portAnnotation, rule.Port, "")
	prefix := param(meta, prefixAnnotation, rule.Prefix, "")
	source := param(meta, sourceAnnotation, rule.Source, nodeName)
	includeLabels := param(meta, labelsAnnotation, rule.IncludeLabels, "true")

	if source == "" {
		source = meta.Name
	}

	name := resourceName(kind, meta)
	port = sanitizePort(meta.Name, port)
	u := baseURL(scheme, ip, port, path, name, source, prefix)
	u = encodeMeta(u, kind, meta)
	u = encodeTags(u, "", rule.Tags)
	if includeLabels == "true" {
		u = encodeTags(u, "label.", meta.Labels)
	}
	u = encodeFilters(u, rule.Filters)
	return u
}

func baseURL(scheme, ip, port, path, name, source, prefix string) string {
	if port != "" {
		port = fmt.Sprintf(":%s", port)
	}
	base := fmt.Sprintf("?url=%s://%s%s%s&name=%s&discovered=true", scheme, ip, port, path, name)
	if source != "" {
		base = fmt.Sprintf("%s&source=%s", base, source)
	}
	if prefix != "" {
		base = fmt.Sprintf("%s&prefix=%s", base, prefix)
	}
	return base
}

func encodeMeta(urlStr, kind string, meta metav1.ObjectMeta) string {
	return fmt.Sprintf("%s&tag=%s:%s&tag=namespace:%s", urlStr, kind, meta.Name, meta.Namespace)
}

func encodeTags(urlStr, prefix string, tags map[string]string) string {
	if len(tags) == 0 {
		return urlStr
	}
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		// exclude pod-template-hash
		if k != "pod-template-hash" {
			urlStr = fmt.Sprintf("%s&tag=%s%s:%s", urlStr, prefix, k, tags[k])
		}
	}
	return urlStr
}

func encodeFilters(urlStr string, cfg filter.Config) string {
	if cfg.Empty() {
		return urlStr
	}
	urlStr = encodeFilter(urlStr, filter.MetricWhitelist, cfg.MetricWhitelist)
	urlStr = encodeFilter(urlStr, filter.MetricBlacklist, cfg.MetricBlacklist)
	urlStr = encodeFilterMap(urlStr, filter.MetricTagWhitelist, cfg.MetricTagWhitelist)
	urlStr = encodeFilterMap(urlStr, filter.MetricTagBlacklist, cfg.MetricTagBlacklist)
	urlStr = encodeFilter(urlStr, filter.TagInclude, cfg.TagInclude)
	urlStr = encodeFilter(urlStr, filter.TagExclude, cfg.TagExclude)
	return urlStr
}

func encodeFilterMap(urlStr, name string, filters map[string][]string) string {
	if len(filters) == 0 {
		return urlStr
	}
	var keys []string
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		patterns := "[" + strings.Join(filters[k], ",") + "]"
		urlStr = fmt.Sprintf("%s&%s=%s:%s", urlStr, name, k, patterns)
	}
	return urlStr
}

func encodeFilter(urlStr, name string, slice []string) string {
	for _, val := range slice {
		urlStr = fmt.Sprintf("%s&%s=%s", urlStr, name, val)
	}
	return urlStr
}

func param(meta metav1.ObjectMeta, annotation, cfgVal, defaultVal string) string {
	// give precedence to annotation
	value := meta.GetAnnotations()[annotation]
	if value == "" {
		// then config
		value = cfgVal
	}
	if value == "" {
		// then default
		value = defaultVal
	}
	return value
}

func resourceName(kind string, meta metav1.ObjectMeta) string {
	if kind == discovery.ServiceType.String() {
		return meta.Namespace + "-" + kind + "-" + meta.Name
	}
	return kind + "-" + meta.Name
}

func sanitizePort(name, port string) string {
	if strings.Contains(name, "kube-state-metrics") && port == "" {
		glog.V(5).Infof("using port 8080 for %s", name)
		return "8080"
	}
	return port
}
