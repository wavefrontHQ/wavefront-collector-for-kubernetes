// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBaseURL(t *testing.T) {
	result := configuration.PrometheusSourceConfig{}
	encodeBase(&result, "http", "192.168.0.1", "9102", "/metrics", "test", "test_source", "test.")

	expected := fmt.Sprintf("%s://%s%s%s", "http", "192.168.0.1", ":9102", "/metrics")
	assert.Equal(t, expected, result.URL)
	assert.Equal(t, "test_source", result.Source)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, "test.", result.Prefix)
}

func TestCustomAnnotation(t *testing.T) {
	s := customAnnotation(schemeAnnotationFormat, "wavefront.com")
	assert.Equal(t, "wavefront.com/scheme", s)
}

func TestEncode(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}

	prefix := "wavefront.com"
	encoder := newPrometheusEncoder(prefix)

	// should return nil without pod IP
	name, promCfg, ok := encoder.Encode("", "pod", pod.ObjectMeta, discovery.PluginConfig{})
	if ok {
		t.Errorf("expected empty scrapeURL for %s. actual: %s", promCfg, name)
	}

	pod.Status = v1.PodStatus{
		PodIP: "10.2.3.4",
	}

	// should return nil if empty cfg and no scrape annotation
	name, promCfg, ok = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, discovery.PluginConfig{})
	if ok {
		t.Errorf("expected empty scrapeURL. actual: %s", promCfg)
	}

	// should return nil if scrape annotation is set to false
	pod.Annotations = map[string]string{customAnnotation(scrapeAnnotationFormat, prefix): "false"}
	name, promCfg, ok = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, discovery.PluginConfig{})
	if ok {
		t.Errorf("expected empty scrapeURL. actual: %s", promCfg)
	}

	// expect non-empty when scrape annotation set to true
	pod.Annotations[customAnnotation(scrapeAnnotationFormat, prefix)] = "true"
	name, promCfg, ok = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, discovery.PluginConfig{})
	if !ok {
		t.Error("expected non-empty scrapeURL.")
	}

	// validate all annotations are picked up
	pod.Labels = map[string]string{"key1": "value1"}
	pod.Annotations[customAnnotation(schemeAnnotationFormat, prefix)] = "https"
	pod.Annotations[customAnnotation(pathAnnotationFormat, prefix)] = "/prometheus"
	pod.Annotations[customAnnotation(portAnnotationFormat, prefix)] = "9102"
	pod.Annotations[customAnnotation(prefixAnnotationFormat, prefix)] = "test."
	pod.Annotations[customAnnotation(labelsAnnotationFormat, prefix)] = "false"
	pod.Annotations[customAnnotation(timeoutAnnotationFormat, prefix)] = "60s"

	name, promCfg, ok = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, discovery.PluginConfig{Name: "test"})
	if !ok {
		t.Error("expected non-empty scrapeURL.")
	}
	pcfg := promCfg.(configuration.PrometheusSourceConfig)

	resName := discovery.ResourceName(discovery.PodType.String(), pod.ObjectMeta)
	assert.Equal(t, fmt.Sprintf("https://%s:9102/prometheus", pod.Status.PodIP), pcfg.URL)
	assert.Equal(t, resName+":9102/prometheus", pcfg.Name)
	assert.Equal(t, "rule", pcfg.Discovered)
	assert.Equal(t, "test", pcfg.Source)
	assert.Equal(t, "test.", pcfg.Prefix)
	assert.Equal(t, 60*time.Second, pcfg.Collection.Timeout)
	checkTag(pcfg.Tags, "pod", "test", t)
	checkTag(pcfg.Tags, "namespace", "test", t)

	// validate cfg is picked up
	cfg := discovery.PluginConfig{
		Name:          "test",
		Scheme:        "https",
		Path:          "/path",
		Port:          "9103",
		Prefix:        "foo.",
		IncludeLabels: "false",
		Conf: `
bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token'
tls_config:
 ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'
 insecure_skip_verify: true
`,
	}
	pod.Annotations = map[string]string{}

	name, promCfg, ok = encoder.Encode("10.2.3.4", "pod", pod.ObjectMeta, cfg)
	pcfg = promCfg.(configuration.PrometheusSourceConfig)

	assert.Equal(t, fmt.Sprintf("https://%s:9103/path", pod.Status.PodIP), pcfg.URL)
	assert.Equal(t, resName+":9103/path", pcfg.Name)
	assert.Equal(t, "rule", pcfg.Discovered)
	assert.Equal(t, "test", pcfg.Source)
	assert.Equal(t, "foo.", pcfg.Prefix)
	checkTag(pcfg.Tags, "pod", "test", t)
	checkTag(pcfg.Tags, "namespace", "test", t)

	assert.Equal(t, "/var/run/secrets/kubernetes.io/serviceaccount/token", pcfg.HTTPClientConfig.BearerTokenFile)
	assert.Equal(t, "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt", pcfg.HTTPClientConfig.TLSConfig.CAFile)
	assert.True(t, pcfg.HTTPClientConfig.TLSConfig.InsecureSkipVerify)
}

func checkTag(tags map[string]string, key, val string, t *testing.T) {
	if len(tags) == 0 {
		t.Error("missing tags")
	}
	if v, ok := tags[key]; ok {
		if v == val {
			return
		}
	}
	t.Errorf("missing tag: %s", key)
}
