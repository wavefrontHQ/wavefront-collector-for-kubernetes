// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/options"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	gm "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
)

func TestNoFilters(t *testing.T) {
	src := &prometheusMetricsSource{}

	points, err := src.parseMetrics(testMetricReader())
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 8, len(points), "wrong number of points")
}

func TestMetricAllowList(t *testing.T) {
	cfg := filter.Config{
		MetricAllowList: []string{"*seconds.count*"},
	}
	f := filter.FromConfig(cfg)

	src := &prometheusMetricsSource{
		filters: f,
	}

	points, err := src.parseMetrics(testMetricReader())
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 1, len(points), "wrong number of points")
}

func TestMetricDenyList(t *testing.T) {
	cfg := filter.Config{
		MetricDenyList: []string{"*seconds.count*"},
	}
	f := filter.FromConfig(cfg)

	src := &prometheusMetricsSource{
		filters: f,
	}

	points, err := src.parseMetrics(testMetricReader())
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 7, len(points), "wrong number of points")
}

func TestMetricTagAllowList(t *testing.T) {
	cfg := filter.Config{
		MetricTagAllowList: map[string][]string{"label": {"good"}},
	}
	f := filter.FromConfig(cfg)

	src := &prometheusMetricsSource{
		filters: f,
	}

	points, err := src.parseMetrics(testMetricReader())
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 1, len(points), "wrong number of points")
}

func TestMetricTagDenyList(t *testing.T) {
	cfg := filter.Config{
		MetricTagDenyList: map[string][]string{"label": {"ba*"}},
	}
	f := filter.FromConfig(cfg)

	src := &prometheusMetricsSource{
		filters: f,
	}

	points, err := src.parseMetrics(testMetricReader())
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 7, len(points), "wrong number of points")
}

func TestTagInclude(t *testing.T) {
	src := &prometheusMetricsSource{
		filters: filter.FromConfig(filter.Config{
			TagInclude: []string{"label"},
		}),
	}

	points, err := src.parseMetrics(testMetricReader())
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 8, len(points), "wrong number of points")

	tagCounts := map[string]int{}
	for _, point := range points {
		tags := point.Tags()
		for tagName := range tags {
			tagCounts[tagName] += 1
		}
	}
	assert.Equal(t, 1, len(tagCounts), "the only tags left are 'label'")
	assert.Equal(t, 2, tagCounts["label"], "two metrics have a tag named 'label'")
}

func TestTagExclude(t *testing.T) {
	src := &prometheusMetricsSource{
		filters: filter.FromConfig(filter.Config{
			TagExclude: []string{"label"},
		}),
	}

	points, err := src.parseMetrics(testMetricReader())
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 8, len(points), "wrong number of points")

	for _, point := range points {
		_, ok := point.Tags()["label"]
		assert.False(t, ok, point.Tags())
	}
}

func BenchmarkMetricPoint(b *testing.B) {
	filtered := gm.GetOrRegisterCounter("filtered", gm.DefaultRegistry)
	tempTags := map[string]string{"pod_name": "prometheus_pod_xyz", "namespace_name": "default"}
	src := &prometheusMetricsSource{prefix: "prefix."}
	pointBuilder := NewPointBuilder(src, filtered)
	for i := 0; i < b.N; i++ {
		_ = pointBuilder.point("http.requests.total.count", 1.0, 0, "", tempTags)
	}
}

func testMetricReader() *bytes.Reader {
	metricsStr := `
http_request_duration_seconds_bucket{le="0.5"} 0
http_request_duration_seconds_bucket{le="1"} 1
http_request_duration_seconds_bucket{le="2"} 2
http_request_duration_seconds_bucket{le="3"} 3
http_request_duration_seconds_bucket{le="5"} 3
http_request_duration_seconds_bucket{le="+Inf"} 3
http_request_duration_seconds_sum{label="bad"} 6
http_request_duration_seconds_count{label="good"} 3
`
	return bytes.NewReader([]byte(metricsStr))
}

func TestDiscoveredPrometheusMetricSource(t *testing.T) {
	t.Run("static source", func(t *testing.T) {
		ms, err := NewPrometheusMetricsSource("", "", "", "", map[string]string{}, nil, httputil.ClientConfig{})

		assert.Nil(t, err)
		assert.False(t, ms.AutoDiscovered(), "prometheus auto-discovery")
	})

	t.Run("discovered source", func(t *testing.T) {
		ms, err := NewPrometheusMetricsSource("", "", "", "some-discovery-method", map[string]string{}, nil, httputil.ClientConfig{})

		assert.Nil(t, err)
		assert.True(t, ms.AutoDiscovered(), "prometheus auto-discovery")
	})
}

func Test_prometheusMetricsSource_Scrape(t *testing.T) {
	t.Run("returns a result with current timestamp", func(t *testing.T) {
		nowTime := time.Now()
		// https://medium.com/zus-health/mocking-outbound-http-requests-in-go-youre-probably-doing-it-wrong-60373a38d2aa
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		promMetSource := prometheusMetricsSource{
			metricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			client:     &http.Client{},
			pps:        gm.NewCounter(),
		}

		result, err := promMetSource.Scrape()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, result.Timestamp, nowTime)
	})

	t.Run("return an error and increments error counters if client fails to get metrics URL", func(t *testing.T) {
		promMetSource := &prometheusMetricsSource{
			metricsURL: "fake metrics URL",
			client:     &http.Client{},
			eps:        gm.NewCounter(),
		}
		collectErrors.Clear()

		_, scrapeError := promMetSource.Scrape()

		assert.NotNil(t, scrapeError)
		assert.Equal(t, int64(1), collectErrors.Count())
		assert.Equal(t, int64(1), promMetSource.eps.Count())
	})

	t.Run("gets the metrics URL", func(t *testing.T) {
		requestedPath := ""
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requestedPath = request.URL.Path
			writer.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		promMetSource := prometheusMetricsSource{
			metricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			client:     &http.Client{},
			pps:        gm.NewCounter(),
		}

		_, err := promMetSource.Scrape()
		assert.NoError(t, err)

		assert.Equal(t, "/fake/metrics/path", requestedPath)
	})

	t.Run("returns an HTTPError and increments error counters on resp error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()
		promMetSource := &prometheusMetricsSource{
			metricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			client:     &http.Client{},
			eps:        gm.NewCounter(),
		}
		expectedErr := &HTTPError{
			MetricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			Status:     "400 Bad Request",
			StatusCode: http.StatusBadRequest,
		}
		collectErrors.Clear()

		_, scrapeError := promMetSource.Scrape()

		assert.Equal(t, expectedErr, scrapeError)
		assert.Equal(t, int64(1), collectErrors.Count())
		assert.Equal(t, int64(1), promMetSource.eps.Count())
	})

	t.Run("returns metrics based on response body and counts number of points", func(t *testing.T) {
		startTimestamp := time.Now().Unix()
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(`
fake_metric{} 1
fake_metric{} 1
`))
		}))
		defer server.Close()

		promMetSource := prometheusMetricsSource{
			metricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			client:     &http.Client{},
			pps:        gm.NewCounter(),
		}

		expectedMetric := wf.NewPoint(
			"fake.metric.value",
			1.0,
			startTimestamp, // not really though
			"",
			nil,
		)
		expectedMetric.SetLabelPairs([]wf.LabelPair{})

		collectedPointsBefore := collectedPoints.Count()
		result, err := promMetSource.Scrape()
		assert.NoError(t, err)
		collectedPointsAfter := collectedPoints.Count()
		assert.Len(t, result.Metrics, 2)
		assert.Equal(t, expectedMetric.Metric, result.Metrics[0].(*wf.Point).Metric)
		assert.Equal(t, expectedMetric.Value, result.Metrics[0].(*wf.Point).Value)
		assert.LessOrEqual(t, expectedMetric.Timestamp, result.Metrics[0].(*wf.Point).Timestamp)
		assert.Equal(t, expectedMetric.Source, result.Metrics[0].(*wf.Point).Source)
		assert.Equal(t, expectedMetric.Tags(), result.Metrics[0].(*wf.Point).Tags())

		assert.Equal(t, int64(2), collectedPointsAfter-collectedPointsBefore)
		assert.Equal(t, int64(2), promMetSource.pps.Count())
	})
}

func Test_prometheusProvider_GetMetricsSources(t *testing.T) {
	t.Run("when use leader election is enabled", func(t *testing.T) {
		t.Run("returns sources when we are the leader", func(t *testing.T) {
			expectedSources := []metrics.Source{&prometheusMetricsSource{
				metricsURL: "fake metrics url",
			}}
			promProvider := prometheusProvider{
				useLeaderElection: true,
				sources:           expectedSources,
			}
			util.SetAgentType(options.AllAgentType)
			leadership.SetLeading(true)
			defer leadership.SetLeading(false)

			sources := promProvider.GetMetricsSources()

			assert.Equal(t, expectedSources, sources)
		})

		t.Run("does not return sources when we are not the leader", func(t *testing.T) {
			expectedSources := []metrics.Source{&prometheusMetricsSource{
				metricsURL: "fake metrics url",
			}}
			promProvider := prometheusProvider{
				useLeaderElection: true,
				sources:           expectedSources,
			}
			util.SetAgentType(options.AllAgentType)

			sources := promProvider.GetMetricsSources()

			assert.Empty(t, sources)
		})
	})

	t.Run("returns sources when use leader election is disabled", func(t *testing.T) {
		expectedSources := []metrics.Source{&prometheusMetricsSource{
			metricsURL: "fake metrics url",
		}}
		promProvider := prometheusProvider{
			useLeaderElection: false,
			sources:           expectedSources,
		}
		util.SetAgentType(options.AllAgentType)

		sources := promProvider.GetMetricsSources()

		assert.Equal(t, expectedSources, sources)
	})
}

func TestNewPrometheusProvider(t *testing.T) {
	t.Run("errors if prometheus URL is missing", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{}
		prometheusProvider, err := NewPrometheusProvider(cfg)
		assert.Nil(t, prometheusProvider)
		assert.NotNil(t, err)
	})

	t.Run("use configured source as source tag", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{
			URL: "fake url",
			Transforms: configuration.Transforms{
				Source: "fake source",
			},
			UseLeaderElection: true,
		}

		leadership.SetLeading(true)
		util.SetAgentType(options.AllAgentType)

		promProvider, err := NewPrometheusProvider(cfg)
		assert.NoError(t, err)
		source := promProvider.GetMetricsSources()[0].(*prometheusMetricsSource)
		assert.Equal(t, "fake source", source.source)
	})

	t.Run("use node name as source tag", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{URL: "fake url"}
		_ = os.Setenv(util.NodeNameEnvVar, "fake node name")
		defer os.Unsetenv(util.NodeNameEnvVar)

		promProvider, err := NewPrometheusProvider(cfg)

		assert.NoError(t, err)
		source := promProvider.GetMetricsSources()[0].(*prometheusMetricsSource)
		assert.Equal(t, "fake node name", source.source)
	})

	t.Run("use 'prom_source' as source tag", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{URL: "fake url"}

		leadership.SetLeading(true)
		util.SetAgentType(options.AllAgentType)

		promProvider, err := NewPrometheusProvider(cfg)
		assert.NoError(t, err)
		source := promProvider.GetMetricsSources()[0].(*prometheusMetricsSource)
		assert.Equal(t, "prom_source", source.source)
	})

	t.Run("default name to URL if not configured", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{URL: "http://test-prometheus-url.com"}

		prometheusProvider, err := NewPrometheusProvider(cfg)

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%s: http://test-prometheus-url.com", providerName), prometheusProvider.Name())
	})

	t.Run("uses configured provider name", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{Name: "fake name", URL: "http://test-prometheus-url.com"}

		prometheusProvider, err := NewPrometheusProvider(cfg)

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%s: fake name", providerName), prometheusProvider.Name())
	})

	t.Run("sources default to not auto discovered", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",
		}

		promProvider, err := NewPrometheusProvider(cfg)

		assert.NoError(t, err)
		source := promProvider.GetMetricsSources()[0].(*prometheusMetricsSource)
		assert.False(t, source.AutoDiscovered())
	})

	t.Run("configures sources to auto discovered", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{
			URL:        "http://test-prometheus-url.com",
			Discovered: "fake discovered",
		}

		promProvider, err := NewPrometheusProvider(cfg)

		assert.NoError(t, err)
		source := promProvider.GetMetricsSources()[0].(*prometheusMetricsSource)
		assert.True(t, source.AutoDiscovered())
	})

	t.Run("metrics source defaults with minimal configuration", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",
		}
		leadership.SetLeading(true)
		util.SetAgentType(options.AllAgentType)

		promProvider, _ := NewPrometheusProvider(cfg)

		source := promProvider.GetMetricsSources()[0].(*prometheusMetricsSource)
		assert.Equal(t, time.Second*30, source.client.Timeout)
		assert.NotNil(t, source.client.Transport)
		assert.Equal(t, "", source.prefix)
		assert.Equal(t, map[string]string(nil), source.tags)
		assert.Equal(t, nil, source.filters)
		assert.Equal(t, "http://test-prometheus-url.com", source.metricsURL)
	})

	t.Run("returns an error if metrics source creation fails", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",
			HTTPClientConfig: httputil.ClientConfig{
				TLSConfig: httputil.TLSConfig{
					KeyFile:            "sldlfdldldfkjkjlfd",
					InsecureSkipVerify: false,
				},
			},
		}

		_, err := NewPrometheusProvider(cfg)

		assert.NotNil(t, err)
	})

	t.Run("prometheus provider sources contains whatever is returned by metrics source constructor", func(t *testing.T) {
		cfg := configuration.PrometheusSourceConfig{
			URL: "fake metrics source url",
			Transforms: configuration.Transforms{
				Source: "fake metrics source source",
				Prefix: "fake metrics source prefix",
			},

			UseLeaderElection: false,
			Discovered:        "fake discovered",
		}
		util.SetAgentType(options.ClusterAgentType)

		prometheusProvider, _ := NewPrometheusProvider(cfg)

		source := prometheusProvider.GetMetricsSources()[0].(*prometheusMetricsSource)
		assert.Equal(t, "fake metrics source url", source.metricsURL)
		assert.Equal(t, "fake metrics source prefix", source.prefix)
		assert.Equal(t, "fake metrics source source", source.source)
	})

	t.Run("when Discovered is present", func(t *testing.T) {
		t.Run("does not use leader election UseLeaderElection is false", func(t *testing.T) {
			cfg := configuration.PrometheusSourceConfig{
				URL:               "http://test-prometheus-url.com",
				UseLeaderElection: false,
				Discovered:        "fake discovered",
			}

			promProvider, _ := NewPrometheusProvider(cfg)

			assert.False(t, promProvider.(*prometheusProvider).useLeaderElection)
		})

		t.Run("uses leader election when UseLeaderElection is true", func(t *testing.T) {
			cfg := configuration.PrometheusSourceConfig{
				URL:               "http://test-prometheus-url.com",
				UseLeaderElection: true,
				Discovered:        "fake discovered",
			}

			promProvider, err := NewPrometheusProvider(cfg)

			assert.NoError(t, err)
			assert.True(t, promProvider.(*prometheusProvider).useLeaderElection)
		})
	})

	t.Run("when Discovered is empty", func(t *testing.T) {
		t.Run("uses leader election even when UseLeaderElection is false", func(t *testing.T) {
			cfg := configuration.PrometheusSourceConfig{
				URL:               "http://test-prometheus-url.com",
				UseLeaderElection: false,
				Discovered:        "",
			}

			promProvider, _ := NewPrometheusProvider(cfg)

			assert.True(t, promProvider.(*prometheusProvider).useLeaderElection)
		})

		t.Run("uses leader election even when UseLeaderElection is true", func(t *testing.T) {
			cfg := configuration.PrometheusSourceConfig{
				URL:               "http://test-prometheus-url.com",
				UseLeaderElection: true,
				Discovered:        "",
			}

			promProvider, _ := NewPrometheusProvider(cfg)

			assert.True(t, promProvider.(*prometheusProvider).useLeaderElection)
		})
	})
}
