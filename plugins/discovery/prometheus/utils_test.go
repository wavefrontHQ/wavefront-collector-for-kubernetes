package prometheus

import (
	"fmt"
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBaseURL(t *testing.T) {
	base := baseURL("http", "192.168.0.1", "9102", "/metrics", "test", "test_source", "test.")
	expected := fmt.Sprintf("?url=%s://%s%s%s&name=%s&discovered=true&source=%s&prefix=%s", "http", "192.168.0.1", ":9102", "/metrics", "test", "test_source", "test.")
	if base != expected {
		t.Errorf("invalid baseURL. expected=%s actual=%s", expected, base)
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
	resName := discovery.ResourceName(discovery.PodType.String(), pod.ObjectMeta)
	expected := fmt.Sprintf("?url=https://%s:9102/prometheus&name=%s&discovered=true&source=test&prefix=test.&tag=pod:test&tag=namespace:test", pod.Status.PodIP, resName)
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
	expected = fmt.Sprintf("?url=https://%s:9103/path&name=%s&discovered=true&source=test&prefix=foo.&tag=pod:test&tag=namespace:test", pod.Status.PodIP, resName)

	if actual != expected {
		t.Errorf("cfg not encoded. expected: %s actual: %s", expected, actual)
	}
}
