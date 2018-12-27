package prometheus

import (
	"fmt"
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBaseURL(t *testing.T) {
	base := baseURL("http", "192.168.0.1", "9102", "/metrics", "test", "test.")
	expected := fmt.Sprintf("?url=%s://%s%s%s&name=%s&prefix=%s", "http", "192.168.0.1", ":9102", "/metrics", "test", "test.")
	if base != expected {
		t.Errorf("invalid baseURL. expected=%s actual=%s", expected, base)
	}
}

func TestEncodeTags(t *testing.T) {
	labels := make(map[string]string)
	labels["a"] = "a"
	labels["b"] = "b"
	encoded := encodeTags("testUrl", labels)
	expected := "testUrl&tag=a:a&tag=b:b"
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
	encoded := encodePod("testUrl", &pod)
	expected := "testUrl&tag=pod:test&tag=namespace:test-ns"
	if encoded != expected {
		t.Errorf("invalid encodePod. expected=%s encoded=%s", expected, encoded)
	}
}

func TestParam(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Annotations: map[string]string{"key1": "value1"},
		},
	}
	p := param(&pod, "key1", "cfgValue", "defaultValue")
	if p != "value1" {
		t.Errorf("expected annotation value: %s actual: %s", "value1", p)
	}
	p = param(&pod, "key2", "cfgValue", "defaultValue")
	if p != "cfgValue" {
		t.Errorf("expected cfg value: %s actual: %s", "cfgValue", p)
	}
	p = param(&pod, "key2", "", "defaultValue")
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
	u := scrapeURL(&pod, discovery.PrometheusConfig{}, false)
	if u != "" {
		t.Errorf("expected empty scrapeURL. actual: %s", u)
	}

	pod.Status = v1.PodStatus{
		PodIP: "192.168.0.1",
	}

	// should return nil if checkAnnotation is true and there is no scrape annotation
	u = scrapeURL(&pod, discovery.PrometheusConfig{}, true)
	if u != "" {
		t.Errorf("expected empty scrapeURL. actual: %s", u)
	}

	// should return nil if scrape annotation is set to false
	pod.Annotations = map[string]string{"prometheus.io/scrape": "false"}
	u = scrapeURL(&pod, discovery.PrometheusConfig{}, true)
	if u != "" {
		t.Errorf("expected empty scrapeURL. actual: %s", u)
	}

	// expect non-empty when scrape annotation set to true
	pod.Annotations["prometheus.io/scrape"] = "true"
	u = scrapeURL(&pod, discovery.PrometheusConfig{}, true)
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

	u = scrapeURL(&pod, discovery.PrometheusConfig{}, true)
	if u == "" {
		t.Error("expected non-empty scrapeURL.")
	}
	expected := fmt.Sprintf("?url=https://%s:9102/prometheus&name=test&prefix=test.&tag=pod:test&tag=namespace:test", pod.Status.PodIP)
	actual := u
	if actual != expected {
		t.Errorf("annotations not encoded. expected: %s actual: %s", expected, actual)
	}

	// validate cfg is picked up
	cfg := discovery.PrometheusConfig{
		Scheme:        "https",
		Path:          "/path",
		Port:          "9103",
		Prefix:        "foo.",
		IncludeLabels: "false",
	}
	pod.Annotations = map[string]string{}

	actual = scrapeURL(&pod, cfg, false)
	expected = fmt.Sprintf("?url=https://%s:9103/path&name=test&prefix=foo.&tag=pod:test&tag=namespace:test", pod.Status.PodIP)

	if actual != expected {
		t.Errorf("cfg not encoded. expected: %s actual: %s", expected, actual)
	}
}
