package prometheus

import (
	"fmt"
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

func scrapeURL(ip, kind string, meta metav1.ObjectMeta, rule discovery.PrometheusConfig) string {
	if ip == "" {
		glog.V(5).Infof("missing ip for %s=%s", kind, meta.Name)
		return ""
	}
	scrape := utils.Param(meta, scrapeAnnotation, "", "false")
	if rule.Name == "" && scrape != "true" {
		glog.V(5).Infof("scrape=false for %s=%s annotations=%q", kind, meta.Name, meta.Annotations)
		return ""
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
	u := baseURL(scheme, ip, port, path, name, source, prefix)
	u = utils.EncodeMeta(u, kind, meta)
	u = utils.EncodeTags(u, "", rule.Tags)
	if includeLabels == "true" {
		u = utils.EncodeTags(u, "label.", meta.Labels)
	}
	u = utils.EncodeFilters(u, rule.Filters)
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

func sanitizePort(name, port string) string {
	if strings.Contains(name, "kube-state-metrics") && port == "" {
		glog.V(5).Infof("using port 8080 for %s", name)
		return "8080"
	}
	return port
}
