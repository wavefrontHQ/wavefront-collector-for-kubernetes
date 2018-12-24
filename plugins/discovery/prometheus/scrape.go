package prometheus

import (
	"fmt"
	"net/url"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

func ScrapeURL(pod *v1.Pod, cfg discovery.PrometheusConfig, checkScrapeAnnotation bool) (*url.URL, error) {
	glog.V(4).Infof("podName=%s podIP=%s podNS=%s", pod.Name, pod.Status.PodIP, pod.Namespace)

	ip := pod.Status.PodIP
	if ip == "" {
		return nil, fmt.Errorf("missing pod ip for %s", pod.Name)
	}
	scrape := getParam(pod, "prometheus.io/scrape", "", "false")
	if checkScrapeAnnotation && scrape != "true" {
		glog.Info("scrape annotation false for pod=", pod.Name, " annotations=", pod.Annotations)
		return nil, nil
	}

	scheme := getParam(pod, "prometheus.io/scheme", cfg.Scheme, "http")
	path := getParam(pod, "prometheus.io/path", cfg.Path, "/metrics")
	port := getParam(pod, "prometheus.io/port", cfg.Port, "")
	prefix := getParam(pod, "prometheus.io/prefix", cfg.Prefix, "")
	includeLabels := getParam(pod, "prometheus.io/includeLabels", cfg.IncludeLabels, "true")

	urlStr := baseURL(scheme, ip, port, path, pod.Name, prefix)
	urlStr = encodePod(urlStr, pod)
	if includeLabels == "true" {
		urlStr = encodeLabels(urlStr, pod.Labels)
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	glog.V(4).Infof("scrapeURL=%s", u)
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

func getParam(pod *v1.Pod, annotation, cfgVal, defaultVal string) string {
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
