// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestImages(t *testing.T) {
	// single image
	rf, err := newResourceFilter(discovery.PluginConfig{
		Type: "telegraf/redis",
		Selectors: discovery.Selectors{
			Images: []string{"redis:*"},
		},
		Port: "6379",
	})
	if err != nil {
		t.Error(err)
	}

	c1 := makeContainer("foobar-redis:1.2.3", []int32{8080})
	c2 := makeContainer("redis:2.8.23", []int32{8080, 6379})

	r1 := makeResource([]v1.Container{c1, c2}, nil, "")
	r2 := makeResource([]v1.Container{c1}, nil, "")

	if !rf.matches(r1) {
		t.Error("container not matching")
	}

	if rf.matches(r2) {
		t.Error("unexpected container match")
	}

	// multiple container images
	rf, err = newResourceFilter(discovery.PluginConfig{
		Type: "telegraf/redis",
		Selectors: discovery.Selectors{
			Images: []string{"redis:*", "*redisslave:v2"},
		},
		Port: "6379",
	})
	if err != nil {
		t.Error(err)
	}

	c3 := makeContainer("gcr.io/google_samples/gb-redisslave:v2", []int32{6379})
	r3 := makeResource([]v1.Container{c3}, nil, "")
	r4 := makeResource([]v1.Container{c2}, nil, "")

	if !rf.matches(r3) {
		t.Errorf("container not matching")
	}
	if !rf.matches(r4) {
		t.Errorf("container not matching")
	}
	if rf.matches(r2) {
		t.Errorf("unexpected container match")
	}

	// image without port
	rf, err = newResourceFilter(discovery.PluginConfig{
		Type: "telegraf/rabbitmq",
		Selectors: discovery.Selectors{
			Images: []string{"rabbitmq:*"},
		},
	})
	if err != nil {
		t.Error(err)
	}
	c3 = makeContainer("rabbitmq:v1.1", []int32{5672})
	r3 = makeResource([]v1.Container{c3}, nil, "")
	if !rf.matches(r3) {
		t.Errorf("container not matching")
	}
}

func TestLabels(t *testing.T) {
	rf, err := newResourceFilter(discovery.PluginConfig{
		Type: "telegraf/redis",
		Selectors: discovery.Selectors{
			Labels: map[string][]string{
				"k8s-app": {"app1-redis*", "*app2-redis"},
			},
		},
		Port: "6379",
	})
	if err != nil {
		t.Error(err)
	}

	r1 := makeResource(nil, nil, "")
	r2 := makeResource(nil, map[string]string{"k8s-app": "app1-redis-db"}, "")
	r3 := makeResource(nil, map[string]string{"k8s-app": "db-app2-redis"}, "")

	if rf.matches(r1) {
		t.Error("invalid label matching")
	}
	if !rf.matches(r2) {
		t.Error("label not matching")
	}
	if !rf.matches(r3) {
		t.Error("label not matching")
	}

	// multiple labels
	rf, err = newResourceFilter(discovery.PluginConfig{
		Type: "telegraf/redis",
		Selectors: discovery.Selectors{
			Labels: map[string][]string{
				"k8s-app": {"app1-redis*", "*app2-redis"},
				"env":     {"dev-1", "dev-2"},
			},
		},
		Port: "6379",
	})
	if err != nil {
		t.Error(err)
	}

	if rf.matches(r1) || rf.matches(r2) {
		t.Error("invalid label matching")
	}
	r4 := makeResource(nil, map[string]string{"k8s-app": "app1-redis", "env": "dev-1"}, "")
	if !rf.matches(r4) {
		t.Error("label not matching")
	}
}

func TestNamespaces(t *testing.T) {
	rf, err := newResourceFilter(discovery.PluginConfig{
		Type: "telegraf/redis",
		Selectors: discovery.Selectors{
			Namespaces: []string{"default*", "collector"},
		},
		Port: "6379",
	})
	if err != nil {
		t.Error(err)
	}

	r1 := makeResource(nil, nil, "default-ns")
	r2 := makeResource(nil, nil, "collector")

	if !rf.matches(r1) {
		t.Error("label not matching")
	}
	if !rf.matches(r2) {
		t.Error("label not matching")
	}

	r3 := makeResource(nil, nil, "collector-ns")
	if rf.matches(r3) {
		t.Error("invalid label matching")
	}
}

func TestAll(t *testing.T) {
	rf, err := newResourceFilter(discovery.PluginConfig{
		Type: "telegraf/redis",
		Selectors: discovery.Selectors{
			Images:     []string{"redis:*", "*redisslave:v2"},
			Namespaces: []string{"default*", "collector"},
			Labels: map[string][]string{
				"k8s-app": {"app1-redis*", "*app2-redis"},
			},
		},
		Port: "6379",
	})
	if err != nil {
		t.Error(err)
	}

	c1 := makeContainer("foobar-redis:1.2.3", []int32{8080})
	c2 := makeContainer("redis:2.8.23", []int32{8080, 6379})

	r1 := makeResource([]v1.Container{c1, c2}, map[string]string{"k8s-app": "app2-redis"}, "default")
	r2 := makeResource([]v1.Container{c1}, nil, "")

	if !rf.matches(r1) {
		t.Error("filter not matching")
	}
	if rf.matches(r2) {
		t.Error("invalid filter matching")
	}
}

func makeResource(containers []v1.Container, labels map[string]string, ns string) discovery.Resource {
	return discovery.Resource{
		Kind:       discovery.PodType.String(),
		Containers: containers,
		Meta: metav1.ObjectMeta{
			Labels:    labels,
			Namespace: ns,
		},
	}
}

func makeContainer(image string, ports []int32) v1.Container {
	c := v1.Container{Image: image}
	c.Ports = make([]v1.ContainerPort, len(ports))
	for i, port := range ports {
		c.Ports[i] = v1.ContainerPort{ContainerPort: port}
	}
	return c
}
