package prometheus

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"testing/quick"
)

type StubNodeLister v1.NodeList

func (s *StubNodeLister) List(_ metav1.ListOptions) (*v1.NodeList, error) {
	return (*v1.NodeList)(s), nil
}

type ErrorNodeLister string

func (s ErrorNodeLister) List(_ metav1.ListOptions) (*v1.NodeList, error) {
	return nil, errors.New(string(s))
}

func TestGenerateConfigs(t *testing.T) {
	nodeLister := &StubNodeLister{Items: []v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "node-2"}},
	}}
	getMyNode := func() string {
		return "node-1"
	}

	t.Run("when the URL has NodeName variable", func(t *testing.T) {
		t.Run("copies every other value unchanged", func(t *testing.T) {
			config := configuration.PrometheusSourceConfig{
				Transforms: configuration.Transforms{
					Source: "foo",
					Prefix: "bar",
				},
				Collection: configuration.CollectionConfig{
					Interval: 1,
					Timeout:  2,
				},
				URL: "https://cluster.local/{{.NodeName}}",
				HTTPClientConfig: httputil.ClientConfig{
					BearerToken:     "123",
					BearerTokenFile: "456.txt",
				},
				PerCluster: true,
				Discovered: "not",
				Name:       "something",
			}
			configs, err := GenerateConfigs(config, nodeLister, getMyNode)

			assert.Nil(t, err)

			for i := range nodeLister.Items {
				actualConfig := configs[i]
				assert.Equal(t, config.Transforms, actualConfig.Transforms)
				assert.Equal(t, config.Collection, actualConfig.Collection)
				assert.Equal(t, config.PerCluster, actualConfig.PerCluster)
				assert.Equal(t, config.Discovered, actualConfig.Discovered)
				assert.Equal(t, config.Name, actualConfig.Name)
			}
		})

		t.Run("interpolates the NodeName variable for each config", func(t *testing.T) {
			f := func(https bool, hostname string, port uint16, prepath string, postpath string) bool {
				scheme := "http"
				if https {
					scheme = "https"
				}
				configs, err := GenerateConfigs(
					configuration.PrometheusSourceConfig{
						URL:        fmt.Sprintf("%s://%s:%d/%s/{{.NodeName}}/%s", scheme, hostname, port, prepath, postpath),
						PerCluster: true,
					},
					nodeLister,
					getMyNode,
				)

				assert.Nil(t, err)

				for i, node := range nodeLister.Items {
					if !assert.Equal(t, fmt.Sprintf("%s://%s:%d/%s/%s/%s", scheme, hostname, port, prepath, node.Name, postpath), configs[i].URL) {
						return false
					}
				}
				return true
			}

			if err := quick.Check(f, nil); err != nil {
				t.Error(err)
			}
		})

		t.Run("returns an error when template cannot parse", func(t *testing.T) {
			_, err := GenerateConfigs(
				configuration.PrometheusSourceConfig{
					URL:        "http://localhost:8080/{{.NodeName}}/prom{{",
					PerCluster: true,
				},
				nodeLister,
				getMyNode,
			)

			assert.Equal(t, "template: :1: unexpected unclosed action in command", err.Error())
		})

		t.Run("returns an error when template cannot parse", func(t *testing.T) {
			_, err := GenerateConfigs(
				configuration.PrometheusSourceConfig{
					URL:        "http://localhost:8080/{{.NodeName}}/prom{{.Foo}}",
					PerCluster: true,
				},
				nodeLister,
				getMyNode,
			)

			assert.Equal(t, "template: :1:42: executing \"\" at <.Foo>: can't evaluate field Foo in type prometheus.urlTemplateEnv", err.Error())
		})

		t.Run("when PerCluster is true", func(t *testing.T) {
			t.Run("produces one config PER node", func(t *testing.T) {
				configs, err := GenerateConfigs(
					configuration.PrometheusSourceConfig{
						URL:        "http://localhost:8080/{{.NodeName}}/prom",
						PerCluster: true,
					},
					nodeLister,
					getMyNode,
				)

				assert.Nil(t, err)
				assert.Equal(t, len(nodeLister.Items), len(configs))
			})

			t.Run("returns an error when it cannot list nodes", func(t *testing.T) {
				expectedErrorStr := "something went wrong"
				_, err := GenerateConfigs(
					configuration.PrometheusSourceConfig{
						URL:        "http://localhost:8080/{{.NodeName}}/prom",
						PerCluster: true,
					},
					ErrorNodeLister(expectedErrorStr),
					getMyNode,
				)

				assert.Equal(t, expectedErrorStr, err.Error())
			})
		})

		t.Run("produces one config when PerCluster is false", func(t *testing.T) {
			configs, err := GenerateConfigs(
				configuration.PrometheusSourceConfig{
					URL:        "http://localhost:8080/{{.NodeName}}/prom",
					PerCluster: false,
				},
				nodeLister,
				getMyNode,
			)

			assert.Nil(t, err)
			assert.Equal(t, 1, len(configs))
		})
	})

	t.Run("when the URL does NOT have a NodeName variable", func(t *testing.T) {
		t.Run("produces one config", func(t *testing.T) {
			configs, err := GenerateConfigs(
				configuration.PrometheusSourceConfig{
					URL:        "http://localhost:8080/some_prom_endpoint",
					PerCluster: true,
				},
				nodeLister,
				getMyNode,
			)

			assert.Nil(t, err)
			assert.Equal(t, 1, len(configs))
		})

		t.Run("copies every other value unchanged", func(t *testing.T) {
			config := configuration.PrometheusSourceConfig{
				Transforms: configuration.Transforms{
					Source: "foo",
					Prefix: "bar",
				},
				Collection: configuration.CollectionConfig{
					Interval: 1,
					Timeout:  2,
				},
				URL: "https://cluster.local/prom",
				HTTPClientConfig: httputil.ClientConfig{
					BearerToken:     "123",
					BearerTokenFile: "456.txt",
				},
				PerCluster: true,
				Discovered: "not",
				Name:       "something",
			}
			configs, err := GenerateConfigs(config, nodeLister, getMyNode)

			assert.Nil(t, err)

			actualConfig := configs[0]
			assert.Equal(t, config.Transforms, actualConfig.Transforms)
			assert.Equal(t, config.Collection, actualConfig.Collection)
			assert.Equal(t, config.PerCluster, actualConfig.PerCluster)
			assert.Equal(t, config.Discovered, actualConfig.Discovered)
			assert.Equal(t, config.Name, actualConfig.Name)

		})
	})
}
