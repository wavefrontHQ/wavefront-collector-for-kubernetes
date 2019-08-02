package utils

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EncodeFilters(values url.Values, cfg filter.Config) {
	//TODO: remove
	if cfg.Empty() {
		return
	}
	encodeFilter(values, filter.MetricWhitelist, cfg.MetricWhitelist)
	encodeFilter(values, filter.MetricBlacklist, cfg.MetricBlacklist)
	encodeFilterMap(values, filter.MetricTagWhitelist, cfg.MetricTagWhitelist)
	encodeFilterMap(values, filter.MetricTagBlacklist, cfg.MetricTagBlacklist)
	encodeFilter(values, filter.TagInclude, cfg.TagInclude)
	encodeFilter(values, filter.TagExclude, cfg.TagExclude)
}

func encodeFilterMap(values url.Values, name string, filters map[string][]string) {
	if len(filters) == 0 {
		return
	}
	var keys []string
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		patterns := "[" + strings.Join(filters[k], ",") + "]"
		values.Add(name, fmt.Sprintf("%s:%s", k, patterns))
	}
}

func encodeFilter(values url.Values, name string, slice []string) {
	//TODO: remove
	for _, val := range slice {
		values.Add(name, val)
	}
}

func EncodeTags(destTags map[string]string, prefix string, tags map[string]string) {
	if len(tags) == 0 {
		return
	}
	for k, v := range tags {
		if k != "pod-template-hash" && len(k) > 0 && len(v) > 0 {
			key := fmt.Sprintf("%s%s", prefix, k)
			destTags[key] = v
		}
	}
}

func EncodeMeta(tags map[string]string, kind string, meta metav1.ObjectMeta) {
	tags[kind] = meta.Name
	if meta.Namespace != "" {
		tags["namespace"] = meta.Namespace
	}
}

func Param(meta metav1.ObjectMeta, annotation, cfgVal, defaultVal string) string {
	value := ""
	// give precedence to annotation
	if annotation != "" {
		value = meta.GetAnnotations()[annotation]
	}
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
