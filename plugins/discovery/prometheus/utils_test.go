package prometheus

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBaseURL(t *testing.T) {
	values := url.Values{}
	encodeBase(values, "http", "192.168.0.1", "9102", "/metrics", "test", "test_source", "test.")

	expected := fmt.Sprintf("%s://%s%s%s", "http", "192.168.0.1", ":9102", "/metrics")
	checkValue(values, "url", expected, t)
	checkValue(values, "source", "test_source", t)
	checkValue(values, "name", "test", t)
	checkValue(values, "prefix", "test.", t)
}

func TestEncode(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}

	encoder := prometheusEncoder{}

	// should return nil without pod IP
	values := encoder.Encode("", "pod", pod.ObjectMeta, discovery.PluginConfig{})
	if len(values) != 0 {
		t.Errorf("expected empty scrapeURL. actual: %s", values)
	}

	pod.Status = v1.PodStatus{
		PodIP: "10.2.3.4",
	}

	// should return nil if empty cfg and no scrape annotation
	values = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, discovery.PluginConfig{})
	if len(values) != 0 {
		t.Errorf("expected empty scrapeURL. actual: %s", values)
	}

	// should return nil if scrape annotation is set to false
	pod.Annotations = map[string]string{"prometheus.io/scrape": "false"}
	values = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, discovery.PluginConfig{})
	if len(values) != 0 {
		t.Errorf("expected empty scrapeURL. actual: %s", values)
	}

	// expect non-empty when scrape annotation set to true
	pod.Annotations["prometheus.io/scrape"] = "true"
	values = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, discovery.PluginConfig{})
	if len(values) == 0 {
		t.Error("expected non-empty scrapeURL.")
	}

	// validate all annotations are picked up
	pod.Labels = map[string]string{"key1": "value1"}
	pod.Annotations[schemeAnnotation] = "https"
	pod.Annotations[pathAnnotation] = "/prometheus"
	pod.Annotations[portAnnotation] = "9102"
	pod.Annotations[prefixAnnotation] = "test."
	pod.Annotations[labelsAnnotation] = "false"

	values = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, discovery.PluginConfig{})
	if len(values) == 0 {
		t.Error("expected non-empty scrapeURL.")
	}
	resName := discovery.ResourceName(discovery.PodType.String(), pod.ObjectMeta)
	checkValue(values, "url", fmt.Sprintf("https://%s:9102/prometheus", pod.Status.PodIP), t)
	checkValue(values, "name", resName, t)
	checkValue(values, "discovered", "rule", t)
	checkValue(values, "source", "test", t)
	checkValue(values, "prefix", "test.", t)
	checkTag(values, "pod:test", t)
	checkTag(values, "namespace:test", t)

	// validate cfg is picked up
	cfg := discovery.PluginConfig{
		Name:          "test",
		Scheme:        "https",
		Path:          "/path",
		Port:          "9103",
		Prefix:        "foo.",
		IncludeLabels: "false",
	}
	pod.Annotations = map[string]string{}

	values = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, cfg)
	checkValue(values, "url", fmt.Sprintf("https://%s:9103/path", pod.Status.PodIP), t)
	checkValue(values, "name", resName, t)
	checkValue(values, "discovered", "rule", t)
	checkValue(values, "source", "test", t)
	checkValue(values, "prefix", "foo.", t)
	checkTag(values, "pod:test", t)
	checkTag(values, "namespace:test", t)
}

func checkTag(values url.Values, val string, t *testing.T) {
	if len(values["tag"]) == 0 {
		t.Error("missing tags")
	}
	tags := values["tag"]
	for _, tag := range tags {
		if tag == val {
			return
		}
	}
	t.Errorf("missing tag: %s", val)
}

func checkValue(values url.Values, name, val string, t *testing.T) {
	if values.Get(name) != val {
		t.Errorf("key:%s expected:%s actual:%s", name, val, values.Get(name))
	}
}
