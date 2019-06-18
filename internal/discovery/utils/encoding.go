package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EncodeFilters(urlStr string, cfg filter.Config) string {
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

func EncodeTags(urlStr, prefix string, tags map[string]string) string {
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

func EncodeMeta(urlStr, kind string, meta metav1.ObjectMeta) string {
	return fmt.Sprintf("%s&tag=%s:%s&tag=namespace:%s", urlStr, kind, meta.Name, meta.Namespace)
}

func Param(meta metav1.ObjectMeta, annotation, cfgVal, defaultVal string) string {
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
