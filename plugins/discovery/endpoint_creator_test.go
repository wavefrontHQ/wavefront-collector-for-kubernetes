// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"testing"

	"github.com/gobwas/glob"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/discovery/prometheus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/discovery/telegraf"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeDummyProviders(handler metrics.ProviderHandler) map[string]discovery.ProviderInfo {
	providers := make(map[string]discovery.ProviderInfo, 2)
	providers["prometheus"] = prometheus.NewProviderInfo(handler, "prom")
	providers["telegraf"] = telegraf.NewProviderInfo(handler)
	return providers
}

func Test_endpointCreator_discoverEndpoints_annotations_dont_happen_when_disabled(t *testing.T) {
	e := &endpointCreator{
		delegates:                  nil,
		providers:                  makeDummyProviders(util.NewDummyProviderHandler(1)),
		disableAnnotationDiscovery: true,
	}

	resource := discovery.Resource{
		Kind:       discovery.PodType.String(),
		IP:         "0.0.0.0",
		Meta:       metav1.ObjectMeta{},
		Containers: make([]v1.Container, 0),
	}

	got := e.discoverEndpoints(resource)
	assert.Equal(t, 0, len(got))
}

func Test_endpointCreator_discoverEndpoints_annotations_happen_when_not_disabled(t *testing.T) {
	e := &endpointCreator{
		delegates:                  nil,
		providers:                  makeDummyProviders(util.NewDummyProviderHandler(1)),
		disableAnnotationDiscovery: false,
	}

	resource := discovery.Resource{
		Kind: discovery.PodType.String(),
		IP:   "0.0.0.0",
		Meta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"prom/scrape": "true",
				"prom/scheme": "https",
				"prom/port":   "8443",
				"prom/path":   "/healthmetrics",
			},
		},
		Containers: make([]v1.Container, 0),
	}

	got := e.discoverEndpoints(resource)
	assert.Equal(t, 1, len(got))
}

func Test_endpointCreator_discoverEndpoints_annotations(t *testing.T) {
	t.Run("filters based on exclude", func(t *testing.T) {
		e := &endpointCreator{
			providers: makeDummyProviders(util.NewDummyProviderHandler(1)),
			annotationExcludes: []*resourceFilter{
				{
					kind:   "pod",
					images: glob.MustCompile("*istio*"),
				},
				{
					kind: "pod",
					labels: map[string]glob.Glob{
						"foo": glob.MustCompile("bar"),
					},
				},
			},
		}

		resource := makePromResource([]v1.Container{
			makeContainer("some/istio/thing", []int32{80}),
			makeContainer("another/thing", []int32{80}),
		}, nil, "")

		assert.Equal(t, 0, len(e.discoverEndpoints(resource)))

		resource = makePromResource([]v1.Container{makeContainer("some/thing", []int32{80})}, nil, "")

		assert.Equal(t, 1, len(e.discoverEndpoints(resource)))

		resource = makePromResource([]v1.Container{makeContainer("some/thing", []int32{80})}, map[string]string{"foo": "bar"}, "")

		assert.Equal(t, 0, len(e.discoverEndpoints(resource)))
	})
}

func Test_endpointCreator_discoverEndpoints_annotations_happen_by_default(t *testing.T) {
	e := &endpointCreator{
		delegates: nil,
		providers: makeDummyProviders(util.NewDummyProviderHandler(1)),
	}

	resource := discovery.Resource{
		Kind: discovery.PodType.String(),
		IP:   "0.0.0.0",
		Meta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"prom/scrape": "true",
				"prom/scheme": "https",
				"prom/port":   "8443",
				"prom/path":   "/healthmetrics",
			},
		},
		Containers: make([]v1.Container, 0),
	}

	got := e.discoverEndpoints(resource)
	assert.Equal(t, 1, len(got))
}

func Test_endpointCreator_doesnt_discover_endpoints_unless_annotated(t *testing.T) {
	e := &endpointCreator{
		delegates:                  nil,
		providers:                  makeDummyProviders(util.NewDummyProviderHandler(1)),
		disableAnnotationDiscovery: false,
	}

	resource := discovery.Resource{
		Kind:       discovery.PodType.String(),
		IP:         "0.0.0.0",
		Meta:       metav1.ObjectMeta{},
		Containers: make([]v1.Container, 0),
	}

	got := e.discoverEndpoints(resource)
	assert.Equal(t, 0, len(got))
}

func makePromResource(containers []v1.Container, labels map[string]string, ns string) discovery.Resource {
	resource := makeResource(containers, labels, ns)
	resource.IP = "0.0.0.0"
	resource.Meta.Annotations = map[string]string{
		"prom/scrape": "true",
		"prom/scheme": "https",
		"prom/port":   "8443",
		"prom/path":   "/healthmetrics",
	}
	return resource
}
