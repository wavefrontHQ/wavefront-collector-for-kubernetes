package prometheus

import (
	"fmt"
	"sort"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

const (
	scrapeAnnotation = "prometheus.io/scrape"
	schemeAnnotation = "prometheus.io/scheme"
	pathAnnotation   = "prometheus.io/path"
	portAnnotation   = "prometheus.io/port"
	prefixAnnotation = "prometheus.io/prefix"
	labelsAnnotation = "prometheus.io/includeLabels"
)

func scrapeURL(pod *v1.Pod, cfg discovery.PrometheusConfig, checkAnnotation bool) string {
	ip := pod.Status.PodIP
	if ip == "" {
		glog.V(5).Infof("missing pod ip for %s", pod.Name)
		return ""
	}
	scrape := param(pod, scrapeAnnotation, "", "false")
	if checkAnnotation && scrape != "true" {
		glog.V(5).Infof("scrape=false for pod=%s annotations=%q", pod.Name, pod.Annotations)
		return ""
	}

	scheme := param(pod, schemeAnnotation, cfg.Scheme, "http")
	path := param(pod, pathAnnotation, cfg.Path, "/metrics")
	port := param(pod, portAnnotation, cfg.Port, "")
	prefix := param(pod, prefixAnnotation, cfg.Prefix, "")
	includeLabels := param(pod, labelsAnnotation, cfg.IncludeLabels, "true")

	u := baseURL(scheme, ip, port, path, pod.Name, prefix)
	u = encodePod(u, pod)
	if includeLabels == "true" {
		u = encodeLabels(u, pod.Labels)
	}
	return u
}

func baseURL(scheme, ip, port, path, name, prefix string) string {
	if port != "" {
		port = fmt.Sprintf(":%s", port)
	}
	base := fmt.Sprintf("?url=%s://%s%s%s&name=%s", scheme, ip, port, path, name)
	if prefix != "" {
		base = fmt.Sprintf("%s&prefix=%s", base, prefix)
	}
	return base
}

func encodePod(urlStr string, pod *v1.Pod) string {
	return fmt.Sprintf("%s&tag=pod:%s&tag=namespace:%s", urlStr, pod.Name, pod.Namespace)
}

func encodeLabels(urlStr string, labels map[string]string) string {
	if len(labels) == 0 {
		return urlStr
	}

	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		// exclude pod-template-hash
		if k != "pod-template-hash" {
			urlStr = fmt.Sprintf("%s&tag=%s:%s", urlStr, k, labels[k])
		}
	}
	return urlStr
}

func param(pod *v1.Pod, annotation, cfgVal, defaultVal string) string {
	// give precedence to pod annotation
	value := pod.GetAnnotations()[annotation]
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
