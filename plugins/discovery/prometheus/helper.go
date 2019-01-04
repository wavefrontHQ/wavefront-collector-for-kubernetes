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

func scrapeURL(ip, resType string, obj metav1.ObjectMeta, cfg discovery.PrometheusConfig, checkAnnotation bool) string {

	if ip == "" {
		glog.V(5).Infof("missing pod ip for %s", obj.Name)
		return ""
	}
	scrape := param(obj, scrapeAnnotation, "", "false")
	if checkAnnotation && scrape != "true" {
		glog.V(5).Infof("scrape=false for pod=%s annotations=%q", obj.Name, obj.Annotations)
		return ""
	}

	scheme := param(obj, schemeAnnotation, cfg.Scheme, "http")
	path := param(obj, pathAnnotation, cfg.Path, "/metrics")
	port := param(obj, portAnnotation, cfg.Port, "")
	prefix := param(obj, prefixAnnotation, cfg.Prefix, "")
	source := param(obj, sourceAnnotation, cfg.Source, "")
	includeLabels := param(obj, labelsAnnotation, cfg.IncludeLabels, "true")

	u := baseURL(scheme, ip, port, path, obj.Name, source, prefix)
	u = encodeMeta(u, resType, obj)
	u = encodeTags(u, cfg.Tags)
	if includeLabels == "true" {
		u = encodeTags(u, obj.Labels)
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

func encodeMeta(urlStr, resType string, obj metav1.ObjectMeta) string {
	return fmt.Sprintf("%s&tag=%s:%s&tag=namespace:%s", urlStr, resType, obj.Name, obj.Namespace)
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

func param(obj metav1.ObjectMeta, annotation, cfgVal, defaultVal string) string {
	// give precedence to annotation
	value := obj.GetAnnotations()[annotation]
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
