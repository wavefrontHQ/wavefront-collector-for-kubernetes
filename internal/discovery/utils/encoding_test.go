package utils

import (
	"fmt"
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEncodeTags(t *testing.T) {
	labels := make(map[string]string)
	labels["a"] = "a"
	labels["b"] = "b"
	encoded := EncodeTags("testUrl", "label.", labels)
	expected := "testUrl&tag=label.a:a&tag=label.b:b"
	if encoded != expected {
		t.Errorf("invalid encodedTags. expected=%s encoded=%s", expected, encoded)
	}
}

func TestEncodePod(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
	}
	encoded := EncodeMeta("testUrl", "pod", pod.ObjectMeta)
	expected := "testUrl&tag=pod:test&tag=namespace:test-ns"
	if encoded != expected {
		t.Errorf("invalid encodeMeta. expected=%s encoded=%s", expected, encoded)
	}
}

func TestEncodeFilter(t *testing.T) {
	encoded := encodeFilter("testUrl", filter.MetricWhitelist, []string{"foo*", "bar*"})
	expected := fmt.Sprintf("testUrl&%s=foo*&%s=bar*", filter.MetricWhitelist, filter.MetricWhitelist)
	if encoded != expected {
		t.Errorf("error encoding filter. expected=%s encoded=%s", expected, encoded)
	}
}

func TestEncodeFilterMap(t *testing.T) {
	actual := encodeFilterMap("testUrl", filter.MetricBlacklist, map[string][]string{
		"env":     {"dev*", "staging*"},
		"cluster": {"*west", "*east"},
	})
	expected := fmt.Sprintf("testUrl&%s=cluster:[*west,*east]&%s=env:[dev*,staging*]", filter.MetricBlacklist, filter.MetricBlacklist)
	if actual != expected {
		t.Error("error encoding filter map")
	}
}

func TestEncodeFilters(t *testing.T) {
	actual := EncodeFilters("testUrl", filter.Config{
		MetricWhitelist:    []string{"kube.dns.http.*"},
		MetricBlacklist:    []string{"kube.dns.probe.*"},
		MetricTagWhitelist: map[string][]string{"env": {"prod*"}},
		MetricTagBlacklist: map[string][]string{"env": {"dev*"}},
		TagInclude:         []string{"cluster"},
		TagExclude:         []string{"pod-template-hash"},
	})
	expected := fmt.Sprintf("testUrl&%s=kube.dns.http.*&%s=kube.dns.probe.*&%s=env:[prod*]&%s=env:[dev*]&%s=cluster&%s=pod-template-hash",
		filter.MetricWhitelist, filter.MetricBlacklist, filter.MetricTagWhitelist, filter.MetricTagBlacklist, filter.TagInclude, filter.TagExclude)
	if actual != expected {
		t.Error("error encoding filters")
	}
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
