// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// TODO prometheus_test for true interface testing but it will break everything
package prometheus

import (
    "bytes"
    "fmt"
    "net/http"
    "testing"

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
    type fields struct {
        metricsURL           string
        prefix               string
        source               string
        tags                 map[string]string
        filters              filter.Filter
        client               *http.Client
        pps                  gm.Counter
        eps                  gm.Counter
        internalMetricsNames []string
        autoDiscovered       bool
        omitBucketSuffix     bool
    }
    tests := []struct {
        name    string
        fields  fields
        want    *metrics.Batch
        wantErr assert.ErrorAssertionFunc
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            src := &prometheusMetricsSource{
                metricsURL:           tt.fields.metricsURL,
                prefix:               tt.fields.prefix,
                source:               tt.fields.source,
                tags:                 tt.fields.tags,
                filters:              tt.fields.filters,
                client:               tt.fields.client,
                pps:                  tt.fields.pps,
                eps:                  tt.fields.eps,
                internalMetricsNames: tt.fields.internalMetricsNames,
                autoDiscovered:       tt.fields.autoDiscovered,
                omitBucketSuffix:     tt.fields.omitBucketSuffix,
            }
            got, err := src.Scrape()
            if !tt.wantErr(t, err, fmt.Sprintf("Scrape()")) {
                return
            }
            assert.Equalf(t, tt.want, got, "Scrape()")
        })
    }
}

func Test_prometheusProvider_GetMetricsSources(t *testing.T) {
    type fields struct {
        DefaultSourceProvider metrics.DefaultSourceProvider
        name                  string
        useLeaderElection     bool
        sources               []metrics.Source
    }
    tests := []struct {
        name   string
        fields fields
        want   []metrics.Source
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := &prometheusProvider{
                DefaultSourceProvider: tt.fields.DefaultSourceProvider,
                name:                  tt.fields.name,
                useLeaderElection:     tt.fields.useLeaderElection,
                sources:               tt.fields.sources,
            }
            assert.Equalf(t, tt.want, p.GetMetricsSources(), "GetMetricsSources()")
        })
    }
}

func TestNewPrometheusProvider(t *testing.T) {
    type args struct {
        cfg configuration.PrometheusSourceConfig
    }
    tests := []struct {
        name    string
        args    args
        want    metrics.SourceProvider
        wantErr assert.ErrorAssertionFunc
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewPrometheusProvider(tt.args.cfg)
            if !tt.wantErr(t, err, fmt.Sprintf("NewPrometheusProvider(%v)", tt.args.cfg)) {
                return
            }
            assert.Equalf(t, tt.want, got, "NewPrometheusProvider(%v)", tt.args.cfg)
        })
    }

    t.Run("errors if prometheus URL is missing", func(t *testing.T) {
        cfg := configuration.PrometheusSourceConfig{}
        prometheusProvider, err := NewPrometheusProvider(cfg)
        assert.Nil(t, prometheusProvider)
        assert.NotNil(t, err)
    })

    t.Run("creates a prometheus provider with name based on configured name or URL", func(t *testing.T) {
        cfg := configuration.PrometheusSourceConfig{
            Name: "test-name",
            URL:  "http://test-prometheus-url.com",
        }
        prometheusProvider, err := NewPrometheusProvider(cfg)
        assert.NoError(t, err)
        assert.Equal(t, fmt.Sprintf("%s: test-name", providerName), prometheusProvider.Name())

        cfg = configuration.PrometheusSourceConfig{
            URL: "http://test-prometheus-url-only.com",
        }
        prometheusProvider, err = NewPrometheusProvider(cfg)
        assert.NoError(t, err)
        assert.Equal(t, fmt.Sprintf("%s: http://test-prometheus-url-only.com", providerName), prometheusProvider.Name())
    })

    t.Run("calls metrics source constructor with correct values", func(t *testing.T) {
        // url, prefix, source, correct discovered value, tags, filters, http config
        fms := fakeMetricsSourceConstructor{}
        cfg := configuration.PrometheusSourceConfig{
            Name: "test-name",
            URL:  "http://test-prometheus-url.com",
        }
        _, err := prometheusProviderWithMetricsSource(fms.newMetricsSource, cfg)
        assert.NoError(t, err)

        assert.Equal(t, fms.metricsURL, "http://test-prometheus-url.com")
        //assert.Equal(t, fms.prefix, )
        //assert.Equal(t, fms.source, )
        //assert.Equal(t, fms.discovered, )
        //assert.Equal(t, fms.tags, )
        //assert.Equal(t, fms.filters, )
        //assert.Equal(t, fms.httpCfg, )
    })

    t.Run("creates a prometheus provider with leader election based on configured leader election or discovery", func(t *testing.T) {

    })

    t.Run("creates a prometheus provider with sources based on config name or URL", func(t *testing.T) {

    })
}

type fakeMetricsSourceConstructor struct {
    metricsURL string
    prefix     string
    source     string
    discovered string
    tags       map[string]string
    filters    filter.Filter
    httpCfg    httputil.ClientConfig
}

func (fms *fakeMetricsSourceConstructor) newMetricsSource(
    metricsURL,
    prefix,
    source,
    discovered string,
    tags map[string]string,
    filters filter.Filter,
    httpCfg httputil.ClientConfig,
) (metrics.Source, error) {
    fms.metricsURL = metricsURL
    fms.prefix = prefix
    fms.source = source
    fms.discovered = discovered
    fms.tags = tags
    fms.filters = filters
    fms.httpCfg = httpCfg

    return nil, nil
}
