package prometheus

import (
	"fmt"
	"sort"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

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
	source := param(meta, sourceAnnotation, rule.Source, "")
	includeLabels := param(meta, labelsAnnotation, rule.IncludeLabels, "true")

	name := resourceName(kind, meta)
	u := baseURL(scheme, ip, port, path, name, source, prefix)
	u = encodeMeta(u, kind, meta)
	u = encodeTags(u, rule.Tags)
	if includeLabels == "true" {
		u = encodeTags(u, meta.Labels)
	}
	return u
}

func baseURL(scheme, ip, port, path, name, source, prefix string) string {
	if port != "" {
		port = fmt.Sprintf(":%s", port)
	}
	base := fmt.Sprintf("?url=%s://%s%s%s&name=%s", scheme, ip, port, path, name)
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

func encodeTags(urlStr string, tags map[string]string) string {
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
			urlStr = fmt.Sprintf("%s&tag=%s:%s", urlStr, k, tags[k])
		}
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
