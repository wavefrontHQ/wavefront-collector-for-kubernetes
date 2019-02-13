package prometheus

import (
	"fmt"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBaseURL(t *testing.T) {
	base := baseURL("http", "192.168.0.1", "9102", "/metrics", "test", "test_source", "test.")
	expected := fmt.Sprintf("?url=%s://%s%s%s&name=%s&source=%s&prefix=%s", "http", "192.168.0.1", ":9102", "/metrics", "test", "test_source", "test.")
	if base != expected {
		t.Errorf("invalid baseURL. expected=%s actual=%s", expected, base)
	}
}

func TestEncodeTags(t *testing.T) {
	labels := make(map[string]string)
	labels["a"] = "a"
	labels["b"] = "b"
	encoded := encodeTags("testUrl", "label.", labels)
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
	encoded := encodeMeta("testUrl", "pod", pod.ObjectMeta)
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
	actual := encodeFilters("testUrl", filter.Config{
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
	p := param(pod.ObjectMeta, "key1", "cfgValue", "defaultValue")
	if p != "value1" {
		t.Errorf("expected annotation value: %s actual: %s", "value1", p)
	}
	p = param(pod.ObjectMeta, "key2", "cfgValue", "defaultValue")
	if p != "cfgValue" {
		t.Errorf("expected cfg value: %s actual: %s", "cfgValue", p)
	}
	p = param(pod.ObjectMeta, "key2", "", "defaultValue")
	if p != "defaultValue" {
		t.Errorf("expected default value: %s actual: %s", "defaultValue", p)
	}
}

func TestScrapeURL(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}

	// should return nil without pod IP
	u := scrapeURL("", "pod", pod.ObjectMeta, discovery.PrometheusConfig{})
	if u != "" {
		t.Errorf("expected empty scrapeURL. actual: %s", u)
	}

	pod.Status = v1.PodStatus{
		PodIP: "10.2.3.4",
	}

	// should return nil if empty cfg and no scrape annotation
	u = scrapeURL("10.2.3.4", "pod", pod.ObjectMeta, discovery.PrometheusConfig{})
	if u != "" {
		t.Errorf("expected empty scrapeURL. actual: %s", u)
	}

	// should return nil if scrape annotation is set to false
	pod.Annotations = map[string]string{"prometheus.io/scrape": "false"}
	u = scrapeURL("10.2.3.4", "pod", pod.ObjectMeta, discovery.PrometheusConfig{})
	if u != "" {
		t.Errorf("expected empty scrapeURL. actual: %s", u)
	}

	// expect non-empty when scrape annotation set to true
	pod.Annotations["prometheus.io/scrape"] = "true"
	u = scrapeURL("10.2.3.4", "pod", pod.ObjectMeta, discovery.PrometheusConfig{})
	if u == "" {
		t.Error("expected non-empty scrapeURL.")
	}

	// validate all annotations are picked up
	pod.Labels = map[string]string{"key1": "value1"}
	pod.Annotations[schemeAnnotation] = "https"
	pod.Annotations[pathAnnotation] = "/prometheus"
	pod.Annotations[portAnnotation] = "9102"
	pod.Annotations[prefixAnnotation] = "test."
	pod.Annotations[labelsAnnotation] = "false"

	u = scrapeURL("10.2.3.4", "pod", pod.ObjectMeta, discovery.PrometheusConfig{})
	if u == "" {
		t.Error("expected non-empty scrapeURL.")
	}
	resName := resourceName(discovery.PodType.String(), pod.ObjectMeta)
	expected := fmt.Sprintf("?url=https://%s:9102/prometheus&name=%s&source=test&prefix=test.&tag=pod:test&tag=namespace:test", pod.Status.PodIP, resName)
	actual := u
	if actual != expected {
		t.Errorf("annotations not encoded. expected: %s actual: %s", expected, actual)
	}

	// validate cfg is picked up
	cfg := discovery.PrometheusConfig{
		Name:          "test",
		Scheme:        "https",
		Path:          "/path",
		Port:          "9103",
		Prefix:        "foo.",
		IncludeLabels: "false",
	}
	pod.Annotations = map[string]string{}

	actual = scrapeURL("10.2.3.4", "pod", pod.ObjectMeta, cfg)
	expected = fmt.Sprintf("?url=https://%s:9103/path&name=%s&source=test&prefix=foo.&tag=pod:test&tag=namespace:test", pod.Status.PodIP, resName)

	if actual != expected {
		t.Errorf("cfg not encoded. expected: %s actual: %s", expected, actual)
	}
}
