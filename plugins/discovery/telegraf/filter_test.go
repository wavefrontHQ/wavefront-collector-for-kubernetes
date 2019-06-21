package telegraf

import (
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	"k8s.io/api/core/v1"
)

func TestFilter(t *testing.T) {
	// single container image
	rf, err := newResourceFilter(discovery.PluginConfig{
		Type:   "telegraf/redis",
		Images: []string{"redis:*"},
		Port:   "6379",
	})

	if err != nil {
		t.Error(err)
	}

	c1 := makeContainer("foobar-redis:1.2.3", []int32{8080})
	c2 := makeContainer("redis:2.8.23", []int32{8080, 6379})

	if !rf.matches(makeResource([]v1.Container{c1, c2})) {
		t.Error("container not matching")
	}

	if rf.matches(makeResource([]v1.Container{c1})) {
		t.Error("unexpected container match")
	}

	// multiple container images
	rf, err = newResourceFilter(discovery.PluginConfig{
		Type:   "telegraf/redis",
		Images: []string{"redis:*", "*redisslave:v2"},
		Port:   "6379",
	})
	if err != nil {
		t.Error(err)
	}

	c3 := makeContainer("gcr.io/google_samples/gb-redisslave:v2", []int32{6379})
	if !rf.matches(makeResource([]v1.Container{c3})) {
		t.Errorf("container not matching")
	}
	if !rf.matches((makeResource([]v1.Container{c2}))) {
		t.Errorf("container not matching")
	}
	if rf.matches(makeResource([]v1.Container{c1})) {
		t.Errorf("unexpected container match")
	}
}

func makeResource(containers []v1.Container) discovery.Resource {
	return discovery.Resource{
		PodSpec: v1.PodSpec{
			Containers: containers,
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
