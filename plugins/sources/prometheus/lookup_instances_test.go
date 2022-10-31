package prometheus

import (
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"strings"
	"testing"
)

const goodHost = "kubernetes.default.svc"

func TestLookupByEndpoints(t *testing.T) {
	t.Run("rejects hosts that do not match a k8s service dns name", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		_, err := LookupByEndpoints(client.CoreV1())("not.a.k8s.host.local")
		require.EqualError(t, err, "host is not a kubernetes service")
	})

	t.Run("returns an error if an endpoints is not found for the service", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		_, err := LookupByEndpoints(client.CoreV1())(goodHost)
		require.ErrorContains(t, err, "not found")
	})

	t.Run("uses an HTTP port", func(t *testing.T) {
		client := fake.NewSimpleClientset(endpoints(func(endpoints *corev1.Endpoints) {
			endpoints.Subsets = []corev1.EndpointSubset{{
				Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}},
				Ports:     []corev1.EndpointPort{{Name: "http", Port: 8080}},
			}}
		}))
		instances, _ := LookupByEndpoints(client.CoreV1())(goodHost)

		require.True(t, strings.HasSuffix(instances[0].Host, ":8080"), "address ends with :8080")
	})

	t.Run("prefers HTTPS ports over HTTP ports", func(t *testing.T) {
		client := fake.NewSimpleClientset(endpoints(func(endpoints *corev1.Endpoints) {
			endpoints.Subsets = []corev1.EndpointSubset{{
				Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}},
				Ports:     []corev1.EndpointPort{{Name: "http", Port: 8080}, {Name: "https", Port: 6443}},
			}}
		}))
		instances, _ := LookupByEndpoints(client.CoreV1())(goodHost)

		require.True(t, strings.HasSuffix(instances[0].Host, ":6443"), "address ends with :6443")
	})

	t.Run("returns error if neither an HTTP or HTTPS port is found", func(t *testing.T) {
		client := fake.NewSimpleClientset(endpoints(func(endpoints *corev1.Endpoints) {
			endpoints.Subsets = []corev1.EndpointSubset{{
				Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}},
				Ports:     []corev1.EndpointPort{{Name: "smtp", Port: 25}},
			}}
		}))
		_, err := LookupByEndpoints(client.CoreV1())(goodHost)

		require.ErrorContains(t, err, "could not find either HTTP or HTTPS port")
	})

	t.Run("returns one instance for every endpoints address", func(t *testing.T) {
		client := fake.NewSimpleClientset(endpoints(func(endpoints *corev1.Endpoints) {
			endpoints.Subsets = []corev1.EndpointSubset{{
				Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}, {IP: "127.0.0.2"}},
				Ports:     []corev1.EndpointPort{{Name: "https", Port: 6443}},
			}}
		}))
		instances, _ := LookupByEndpoints(client.CoreV1())(goodHost)

		require.Len(t, instances, 2)
	})

	t.Run("returns an instance tag that matches the instance's Host", func(t *testing.T) {
		client := fake.NewSimpleClientset(endpoints(defaultSubset))
		instances, _ := LookupByEndpoints(client.CoreV1())(goodHost)

		require.Equal(t, instances[0].Host, instances[0].Tags["instance"])
	})

	t.Run("returns ip address of the endpoints in the instance's Host", func(t *testing.T) {
		client := fake.NewSimpleClientset(endpoints(defaultSubset))
		instances, _ := LookupByEndpoints(client.CoreV1())(goodHost)

		require.True(t, strings.HasPrefix(instances[0].Host, "127.0.0.1:"), "begins with 127.0.0.1:")
	})

	t.Run("returns the union of addresses for every subset", func(t *testing.T) {
		client := fake.NewSimpleClientset(endpoints(func(endpoints *corev1.Endpoints) {
			endpoints.Subsets = []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}, {IP: "127.0.0.2"}},
					Ports:     []corev1.EndpointPort{{Name: "https", Port: 6443}},
				},
				{
					Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}, {IP: "127.0.0.3"}},
					Ports:     []corev1.EndpointPort{{Name: "https", Port: 6443}},
				},
			}
		}))
		instances, _ := LookupByEndpoints(client.CoreV1())(goodHost)

		require.Len(t, instances, 3)
		instanceHosts := hosts(instances)
		require.Contains(t, instanceHosts, "127.0.0.1:6443")
		require.Contains(t, instanceHosts, "127.0.0.2:6443")
		require.Contains(t, instanceHosts, "127.0.0.3:6443")
	})
}

func hosts(instances []Instance) []string {
	var instanceHosts []string
	for _, instance := range instances {
		instanceHosts = append(instanceHosts, instance.Host)
	}
	return instanceHosts
}

func endpoints(options ...func(*corev1.Endpoints)) *corev1.Endpoints {
	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubernetes",
			Namespace: "default",
		},
		Subsets: nil,
	}
	for _, option := range options {
		option(endpoints)
	}
	return endpoints
}

func defaultSubset(endpoints *corev1.Endpoints) {
	endpoints.Subsets = []corev1.EndpointSubset{{
		Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}},
		Ports:     []corev1.EndpointPort{{Name: "https", Port: 6443}},
	}}
}

func withSubsets(subsets ...corev1.EndpointSubset) func(endpoints *corev1.Endpoints) {
	return func(endpoints *corev1.Endpoints) {
		endpoints.Subsets = subsets
	}
}

func subsetIPs(ips ...string) corev1.EndpointSubset {
	subset := corev1.EndpointSubset{
		Addresses: []corev1.EndpointAddress{},
		Ports:     []corev1.EndpointPort{{Name: "https", Port: 6443}},
	}
	for _, ip := range ips {
		subset.Addresses = append(subset.Addresses, corev1.EndpointAddress{IP: ip})
	}
	return subset
}
