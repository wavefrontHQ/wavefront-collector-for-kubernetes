// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// TODO prometheus_test for true interface testing but it will break everything
package prometheus

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/options"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

type ExpectScrapeErrorWithErrorCounts struct {
	scrapeError    error
	errCountBefore int64
	errCountAfter  int64
	promMetSource  *prometheusMetricsSource
}

func (e *ExpectScrapeErrorWithErrorCounts) runScrape(promMetSource *prometheusMetricsSource) {
	e.promMetSource = promMetSource
	e.promMetSource.eps = gm.NewCounter()
	e.errCountBefore = collectErrors.Count()

	_, e.scrapeError = promMetSource.Scrape()

	e.errCountAfter = collectErrors.Count()
}

func (e *ExpectScrapeErrorWithErrorCounts) runScrapeWithParseMetrics(promMetSource *prometheusMetricsSource, parseMetrics metricsParser) {
	e.promMetSource = promMetSource
	e.promMetSource.eps = gm.NewCounter()
	e.errCountBefore = collectErrors.Count()

	_, e.scrapeError = promMetSource.scrapeWithParseMetrics(parseMetrics)

	e.errCountAfter = collectErrors.Count()
}

func (e *ExpectScrapeErrorWithErrorCounts) verifyErrorNotNil(t *testing.T) {
	assert.NotNil(t, e.scrapeError)
}

func (e *ExpectScrapeErrorWithErrorCounts) verifyErrorEquals(t *testing.T, expectedError interface{}) {
	assert.Equal(t, expectedError, e.scrapeError)
}

func (e *ExpectScrapeErrorWithErrorCounts) verifyErrorCountsIncreased(t *testing.T) {
	assert.Equal(t, int64(1), e.errCountAfter-e.errCountBefore)
	assert.Equal(t, int64(1), e.promMetSource.eps.Count())
}

// TODO actual Scrape tests
func Test_prometheusMetricsSource_Scrape(t *testing.T) {
	t.Run("returns a result with current timestamp", func(t *testing.T) {
		nowTime := time.Now()
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		promMetSource := prometheusMetricsSource{
			metricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			client:     &http.Client{},
			pps:        gm.NewCounter(),
		}

		stubParseMetrics := func(reader io.Reader) ([]wf.Metric, error) {
			return nil, nil
		}
		result, err := promMetSource.scrapeWithParseMetrics(stubParseMetrics)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, result.Timestamp, nowTime)
	})

	t.Run("return an error and increments error counters if client fails to get metrics URL", func(t *testing.T) {
		promMetSource := &prometheusMetricsSource{
			metricsURL: "fake metrics URL",
			client:     &http.Client{},
		}

		e := ExpectScrapeErrorWithErrorCounts{}
		e.runScrape(promMetSource)
		e.verifyErrorNotNil(t)
		e.verifyErrorCountsIncreased(t)
	})

	t.Run("gets the metrics URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/fake/metrics/path" {
				t.Errorf("expected request to '/fake/metrics/path', got '%s'", request.URL.Path)
			}

			writer.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		promMetSource := prometheusMetricsSource{
			metricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			client:     &http.Client{},
			pps:        gm.NewCounter(),
		}

		stubParseMetrics := func(reader io.Reader) ([]wf.Metric, error) {
			return nil, nil
		}
		_, err := promMetSource.scrapeWithParseMetrics(stubParseMetrics)
		assert.NoError(t, err)
	})

	// TODO should I test response close?
	// t.Run("closes response body at end of scrape to prevent leaking", func(t *testing.T) {}

	t.Run("returns an HTTPError and increments error counters on resp error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		promMetSource := &prometheusMetricsSource{
			metricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			client:     &http.Client{},
		}

		expectedErr := &HTTPError{
			MetricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			Status:     "400 Bad Request",
			StatusCode: http.StatusBadRequest,
		}

		e := ExpectScrapeErrorWithErrorCounts{}
		e.runScrape(promMetSource)
		e.verifyErrorEquals(t, expectedErr)
		e.verifyErrorCountsIncreased(t)
	})

	t.Run("returns an error and increments error counters if parseMetrics fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		promMetSource := &prometheusMetricsSource{
			metricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			client:     &http.Client{},
		}

		stubParseMetrics := func(reader io.Reader) ([]wf.Metric, error) {
			return nil, errors.New("fake failed to parse metrics")
		}

		e := ExpectScrapeErrorWithErrorCounts{}
		e.runScrapeWithParseMetrics(promMetSource, stubParseMetrics)
		e.verifyErrorNotNil(t)
		e.verifyErrorCountsIncreased(t)
	})

	t.Run("returns metrics based on response body and counts number of points", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("fake metrics"))
		}))
		defer server.Close()

		promMetSource := prometheusMetricsSource{
			metricsURL: fmt.Sprintf("%s/fake/metrics/path", server.URL),
			client:     &http.Client{},
			pps:        gm.NewCounter(),
		}

		stubMetric := wf.NewPoint(
			"fake metric",
			1.0,
			time.Now().Unix(),
			"fake source",
			map[string]string{},
		)
		stubParseMetrics := func(reader io.Reader) ([]wf.Metric, error) {
			return []wf.Metric{
				stubMetric,
				stubMetric,
			}, nil
		}

		collectedPointsBefore := collectedPoints.Count()
		result, err := promMetSource.scrapeWithParseMetrics(stubParseMetrics)
		assert.NoError(t, err)
		collectedPointsAfter := collectedPoints.Count()
		assert.Equal(t, []wf.Metric{stubMetric, stubMetric}, result.Metrics)

		assert.Equal(t, int64(2), collectedPointsAfter-collectedPointsBefore)
		assert.Equal(t, int64(2), promMetSource.pps.Count())
	})
}

func Test_prometheusProvider_GetMetricsSources(t *testing.T) {
	t.Run("returns sources dependent on leadership election and leading status", func(t *testing.T) {
		promProvider := prometheusProvider{
			useLeaderElection: false,
			sources: []metrics.Source{&prometheusMetricsSource{
				metricsURL: "fake metrics url",
			}},
		}

		sources := promProvider.GetMetricsSources()
		assert.Equal(t, []metrics.Source{&prometheusMetricsSource{
			metricsURL: "fake metrics url",
		}}, sources)

		promProvider.useLeaderElection = true
		allAgentType, err := options.NewAgentType("all")
		assert.NoError(t, err)
		util.SetAgentType(allAgentType)
		sources = promProvider.GetMetricsSources()
		assert.Nil(t, sources)

		promProvider.useLeaderElection = true
		clusterAgentType, err := options.NewAgentType("cluster")
		assert.NoError(t, err)
		util.SetAgentType(clusterAgentType)
		sources = promProvider.GetMetricsSources()
		assert.Nil(t, sources)

		promProvider.useLeaderElection = false
		nodeAgentType, err := options.NewAgentType("node")
		assert.NoError(t, err)
		util.SetAgentType(nodeAgentType)
		sources = promProvider.GetMetricsSources()
		assert.Equal(t, []metrics.Source{&prometheusMetricsSource{
			metricsURL: "fake metrics url",
		}}, sources)

		promProvider.useLeaderElection = true
		legacyAgentType, err := options.NewAgentType("legacy")
		assert.NoError(t, err)
		util.SetAgentType(legacyAgentType)
		sources = promProvider.GetMetricsSources()
		assert.Nil(t, sources)
	})
}

func TestNewPrometheusProvider(t *testing.T) {
	t.Run("errors if prometheus URL is missing", func(t *testing.T) {
		mockPDI := mockPrometheusProviderDependencyInjector{}
		cfg := configuration.PrometheusSourceConfig{}
		prometheusProvider, err := prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.Nil(t, prometheusProvider)
		assert.NotNil(t, err)
	})

	t.Run("use configured source, node name, or 'prom_source' as source tag", func(t *testing.T) {
		mockPDI := mockPrometheusProviderDependencyInjector{}
		cfg := configuration.PrometheusSourceConfig{
			URL: "fake url",
			Transforms: configuration.Transforms{
				Source: "fake source",
			},
		}
		_, err := prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)
		assert.Equal(t, "fake source", mockPDI.source)

		mockPDI = mockPrometheusProviderDependencyInjector{}
		cfg = configuration.PrometheusSourceConfig{
			URL: "fake url",
		}
		_, err = prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "fake node name", cfg)
		assert.NoError(t, err)
		assert.Equal(t, "fake node name", mockPDI.source)

		mockPDI = mockPrometheusProviderDependencyInjector{}
		_, err = prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)
		assert.Equal(t, "prom_source", mockPDI.source)
	})

	t.Run("default name to URL if not configured", func(t *testing.T) {
		mockPDI := mockPrometheusProviderDependencyInjector{}
		cfg := configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",
		}
		prometheusProvider, err := prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%s: http://test-prometheus-url.com", providerName), prometheusProvider.Name())

		cfg.Name = "fake name"
		prometheusProvider, err = prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%s: fake name", providerName), prometheusProvider.Name())
	})

	t.Run("default discovered to empty if not configured", func(t *testing.T) {
		mockPDI := mockPrometheusProviderDependencyInjector{}
		cfg := configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",
		}
		_, err := prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)
		assert.Equal(t, "", mockPDI.discovered)

		cfg.Discovered = "fake discovered"
		_, err = prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)
		assert.Equal(t, "fake discovered", mockPDI.discovered)
	})

	t.Run("metrics source defaults with minimal configuration", func(t *testing.T) {
		mockPDI := mockPrometheusProviderDependencyInjector{}
		cfg := configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",
		}
		_, err := prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)

		assert.Equal(t, httputil.ClientConfig{}, mockPDI.httpCfg)
		assert.Equal(t, "", mockPDI.prefix)
		assert.Equal(t, map[string]string(nil), mockPDI.tags)
		assert.Equal(t, nil, mockPDI.filters)

		assert.Equal(t, "http://test-prometheus-url.com", mockPDI.metricsURL)
	})

	t.Run("returns an error if metrics source creation fails", func(t *testing.T) {
		mockPDI := mockPrometheusProviderDependencyInjector{
			returnError: errors.New("fake metrics source error"),
		}
		cfg := configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",
		}
		_, err := prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NotNil(t, err)
	})

	t.Run("prometheus provider sources contains whatever is returned by metrics source constructor", func(t *testing.T) {
		mockPDI := mockPrometheusProviderDependencyInjector{
			returnMetricsSource: &prometheusMetricsSource{
				metricsURL: "fake metrics source url",
				prefix:     "fake metrics source prefix",
				source:     "fake metrics source source",
			},
		}

		// TODO Jackpot. Behold the calamity of code necessary to get this test to pass.
		cfg := configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",

			UseLeaderElection: false,
			Discovered:        "fake discovered",
		}
		prometheusProvider, err := prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)

		mockAgentType, err := options.NewAgentType("cluster")
		assert.NoError(t, err)
		util.SetAgentType(mockAgentType)
		fakeSource := prometheusProvider.GetMetricsSources()[0].(*prometheusMetricsSource)

		// notes from Matt
		// leader election is what makes this have to be global
		// can't dependency inject agent type and still have leader election global
		// lot of subtle implications not well covered by tests
		// increasing test coverage may help rip out global leadership
		// let needs of function drive order; do what feels most obvious
		// first test -> PR -> then try to make the major refactors
		// just the tests alone would be valuable
		// test -> marinate -> see problems holistically -> refactor fearlessly
		// providers are factories for sources; they could be a function that creates a scrape function--crazy refactor though

		// new technique: comment THE WHOLE FUNCTION and TDD the uncommenting

		// what if we used an Observer pattern for leadership election??

		assert.Equal(t, "fake metrics source url", fakeSource.metricsURL)
		assert.Equal(t, "fake metrics source prefix", fakeSource.prefix)
		assert.Equal(t, "fake metrics source source", fakeSource.source)
	})

	t.Run("creates a prometheus provider with leader election based on configured leader election or discovery", func(t *testing.T) {
		mockPDI := mockPrometheusProviderDependencyInjector{}

		cfg := configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",

			UseLeaderElection: false,
			Discovered:        "fake discovered",
		}
		promProvider, err := prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)

		// TODO should not be testing internal fields, but for now this shines a light on
		// questions I have about the design of this.
		assert.False(t, promProvider.(*prometheusProvider).useLeaderElection)

		cfg = configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",

			UseLeaderElection: true,
			Discovered:        "fake discovered",
		}
		promProvider, err = prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)
		assert.True(t, promProvider.(*prometheusProvider).useLeaderElection)

		cfg = configuration.PrometheusSourceConfig{
			URL: "http://test-prometheus-url.com",

			UseLeaderElection: false,
			Discovered:        "",
		}
		promProvider, err = prometheusProviderWithMetricsSource(mockPDI.newMetricsSource, "", cfg)
		assert.NoError(t, err)
		assert.True(t, promProvider.(*prometheusProvider).useLeaderElection)
	})

	t.Run("creates a prometheus provider with sources based on config name or URL", func(t *testing.T) {
		// TODO what is this? What did I intend with this test?
	})
}

type mockPrometheusProviderDependencyInjector struct {
	metricsURL string
	prefix     string
	source     string
	discovered string
	tags       map[string]string
	filters    filter.Filter
	httpCfg    httputil.ClientConfig

	returnError         error
	returnMetricsSource metrics.Source
}

func (pdi *mockPrometheusProviderDependencyInjector) newMetricsSource(
	metricsURL,
	prefix,
	source,
	discovered string,
	tags map[string]string,
	filters filter.Filter,
	httpCfg httputil.ClientConfig,
) (metrics.Source, error) {
	pdi.metricsURL = metricsURL
	pdi.prefix = prefix
	pdi.source = source
	pdi.discovered = discovered
	pdi.tags = tags
	pdi.filters = filters
	pdi.httpCfg = httpCfg

	if pdi.returnError != nil {
		return nil, pdi.returnError
	}

	return pdi.returnMetricsSource, nil
}

//t.Run("sending own status", func(t *testing.T) {
//    stubKM := test_helper.NewMockKubernetesManager()
//    expectStatusSent := test_helper.NewExpectStatusSent(wf.WavefrontStatus{}, "testClusterName")
//
//    // TODO: so much setup for only one usage...
//    r, wfCR, apiClient, _ := setup("testWavefrontUrl", "updatedToken", "testClusterName")
//    r.KubernetesManager = stubKM
//    r.StatusSender = expectStatusSent
//
//    err := apiClient.Delete(context.Background(), wfCR)
//
//    _, err = r.Reconcile(context.Background(), defaultRequest())
//    assert.NoError(t, err)
//
//    expectStatusSent.Verify(t)
//})

//type MockStatusSender struct {
//    lastStatus      wf.WavefrontStatus
//    lastClusterName string
//
//    expectedStatus      wf.WavefrontStatus
//    expectedClusterName string
//}
//
//func NewMockStatusSender(expectedStatus wf.WavefrontStatus, expectedClusterName string) *MockStatusSender {
//    return &MockStatusSender{
//        expectedStatus:      expectedStatus,
//        expectedClusterName: expectedClusterName,
//    }
//}
//
//func (m *MockStatusSender) SendStatus(status wf.WavefrontStatus, clusterName string) error {
//    m.lastStatus = status
//    m.lastClusterName = clusterName
//    return nil
//}
//
//func (m *MockStatusSender) Close() {
//    panic("did not expect Close to be called on MockStatusSender")
//}
//
//func (m *MockStatusSender) Verify(t *testing.T)  {
//    require.Equal(t, m.expectedClusterName, m.lastClusterName)
//    require.Equal(t, m.expectedStatus, m.lastStatus)
//}
