package prometheus

import (
	"bytes"
	"fmt"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
	"testing"

	dto "github.com/prometheus/client_model/go"
)

var result map[string]string

func BenchmarkBuildTags1(b *testing.B)  { benchmarkBuildTags(1, b) }
func BenchmarkBuildTags2(b *testing.B)  { benchmarkBuildTags(2, b) }
func BenchmarkBuildTags4(b *testing.B)  { benchmarkBuildTags(4, b) }
func BenchmarkBuildTags8(b *testing.B)  { benchmarkBuildTags(8, b) }
func BenchmarkBuildTags16(b *testing.B) { benchmarkBuildTags(16, b) }

func benchmarkBuildTags(i int, b *testing.B) {
	var r map[string]string
	p := &prometheusMetricsSource{
		buf:      bytes.NewBufferString(""),
		flatTags: reporting.EncodeKey("", buildSrcTags(i)),
	}
	m := buildPromMetric(i)
	gt := util.NewGroupedTags()
	for n := 0; n < b.N; n++ {
		r = p.buildTags(m, gt)
	}
	result = r
}

func buildSrcTags(count int) map[string]string {
	m := make(map[string]string)
	for i := 0; i < count; i++ {
		k := fmt.Sprintf("src%d", i)
		m[k] = k
	}
	return m
}

func buildPromMetric(count int) *dto.Metric {
	lp := make([]*dto.LabelPair, count)
	for i := 0; i < count; i++ {
		k := fmt.Sprintf("k%d", i)
		lp[i] = &dto.LabelPair{
			Name:  &k,
			Value: &k,
		}
	}
	return &dto.Metric{
		Label: lp,
	}
}
