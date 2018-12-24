package prometheus

import (
	"fmt"
	"net/url"

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

func scrapeURL(pod *v1.Pod, cfg discovery.PrometheusConfig, checkAnnotation bool) (*url.URL, error) {
	ip := pod.Status.PodIP
	if ip == "" {
		return nil, fmt.Errorf("missing pod ip for %s", pod.Name)
	}
	scrape := param(pod, scrapeAnnotation, "", "false")
	if checkAnnotation && scrape != "true" {
		glog.Infof("scrape=false for pod=%s annotations=%q", pod.Name, pod.Annotations)
		return nil, nil
	}

	scheme := param(pod, schemeAnnotation, cfg.Scheme, "http")
	path := param(pod, pathAnnotation, cfg.Path, "/metrics")
	port := param(pod, portAnnotation, cfg.Port, "")
	prefix := param(pod, prefixAnnotation, cfg.Prefix, "")
	includeLabels := param(pod, labelsAnnotation, cfg.IncludeLabels, "true")

	urlStr := baseURL(scheme, ip, port, path, pod.Name, prefix)
	urlStr = encodePod(urlStr, pod)
	if includeLabels == "true" {
		urlStr = encodeLabels(urlStr, pod.Labels)
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	glog.V(5).Infof("scrapeURL=%s", u)
	return u, nil
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
	for k, v := range labels {
		// exclude pod-template-hash
		if k != "pod-template-hash" {
			urlStr = fmt.Sprintf("%s&tag=%s:%s", urlStr, k, v)
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
