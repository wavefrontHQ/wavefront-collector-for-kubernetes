package utils

import (
	"net/url"
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEncodeTags(t *testing.T) {
	labels := make(map[string]string)
	labels["a"] = "a"
	labels["b"] = "b"
	values := url.Values{}
	EncodeTags(values, "label.", labels)
	checkValues(values, "tag", "label.a:a", t)
	checkValues(values, "tag", "label.b:b", t)
}

func TestEncodePod(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
	}
	values := url.Values{}
	EncodeMeta(values, "pod", pod.ObjectMeta)
	checkValues(values, "tag", "pod:test", t)
	checkValues(values, "tag", "namespace:test-ns", t)
}

func TestEncodeFilter(t *testing.T) {
	values := url.Values{}
	encodeFilter(values, filter.MetricWhitelist, []string{"foo*", "bar*"})
	checkValues(values, filter.MetricWhitelist, "foo*", t)
	checkValues(values, filter.MetricWhitelist, "bar*", t)
}

func TestEncodeFilterMap(t *testing.T) {
	values := url.Values{}
	encodeFilterMap(values, filter.MetricBlacklist, map[string][]string{
		"env":     {"dev*", "staging*"},
		"cluster": {"*west", "*east"},
	})
	checkValues(values, filter.MetricBlacklist, "env:[dev*,staging*]", t)
	checkValues(values, filter.MetricBlacklist, "cluster:[*west,*east]", t)
}

func TestEncodeFilters(t *testing.T) {
	values := url.Values{}
	EncodeFilters(values, filter.Config{
		MetricWhitelist:    []string{"kube.dns.http.*"},
		MetricBlacklist:    []string{"kube.dns.probe.*"},
		MetricTagWhitelist: map[string][]string{"env": {"prod*"}},
		MetricTagBlacklist: map[string][]string{"env": {"dev*"}},
		TagInclude:         []string{"cluster"},
		TagExclude:         []string{"pod-template-hash"},
	})
	checkValue(values, filter.MetricWhitelist, "kube.dns.http.*", t)
	checkValue(values, filter.MetricBlacklist, "kube.dns.probe.*", t)
	checkValues(values, filter.MetricTagWhitelist, "env:[prod*]", t)
	checkValues(values, filter.MetricTagBlacklist, "env:[dev*]", t)
	checkValue(values, filter.TagInclude, "cluster", t)
	checkValue(values, filter.TagExclude, "pod-template-hash", t)
}

func TestParam(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Annotations: map[string]string{"key1": "value1"},
		},
	}
	p := Param(pod.ObjectMeta, "key1", "cfgValue", "defaultValue")
	if p != "value1" {
		t.Errorf("expected annotation value: %s actual: %s", "value1", p)
	}
	p = Param(pod.ObjectMeta, "key2", "cfgValue", "defaultValue")
	if p != "cfgValue" {
		t.Errorf("expected cfg value: %s actual: %s", "cfgValue", p)
	}
	p = Param(pod.ObjectMeta, "key2", "", "defaultValue")
	if p != "defaultValue" {
		t.Errorf("expected default value: %s actual: %s", "defaultValue", p)
	}
}

func checkValues(values url.Values, name, val string, t *testing.T) {
	if len(values[name]) == 0 {
		t.Errorf("missing %s", name)
	}
	tags := values[name]
	for _, tag := range tags {
		if tag == val {
			return
		}
	}
	t.Errorf("missing %s: %s", name, val)
}

func checkValue(values url.Values, name, val string, t *testing.T) {
	if values.Get(name) != val {
		t.Errorf("key:%s expected:%s actual:%s", name, val, values.Get(name))
	}
}
