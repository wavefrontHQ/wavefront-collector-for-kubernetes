package utils

import (
	"fmt"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/httputil"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EncodeFilters(values url.Values, cfg filter.Config) {
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
	for _, val := range slice {
		values.Add(name, val)
	}
}

func EncodeTags(values url.Values, prefix string, tags map[string]string) {
	if len(tags) == 0 {
		return
	}
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		// exclude pod-template-hash
		if k != "pod-template-hash" {
			values.Add("tag", fmt.Sprintf("%s%s:%s", prefix, k, tags[k]))
		}
	}
}

func EncodeMeta(values url.Values, kind string, meta metav1.ObjectMeta) {
	values.Add("tag", fmt.Sprintf("%s:%s", kind, meta.Name))
	if meta.Namespace != "" {
		values.Add("tag", fmt.Sprintf("%s:%s", "namespace", meta.Namespace))
	}
}

func EncodeHTTPConfig(values url.Values, cfg httputil.ClientConfig) {
	addVal(values, "bearerToken", cfg.BearerToken)
	addVal(values, "bearerTokenFile", cfg.BearerTokenFile)
	addVal(values, "tlsCAFile", cfg.TLSConfig.CAFile)
	addVal(values, "tlsCertFile", cfg.TLSConfig.CertFile)
	addVal(values, "tlsKeyFile", cfg.TLSConfig.KeyFile)
	addVal(values, "tlsServerName", cfg.TLSConfig.ServerName)
	addVal(values, "tlsInsecure", strconv.FormatBool(cfg.TLSConfig.InsecureSkipVerify))
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

func addVal(values url.Values, key, val string) {
	if val != "" {
		values.Add(key, val)
	}
}
