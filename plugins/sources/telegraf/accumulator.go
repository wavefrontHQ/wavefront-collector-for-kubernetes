// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package telegraf

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/influxdata/telegraf"
	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

// Implements the telegraf Accumulator interface
type telegrafDataBatch struct {
	metrics.DataBatch
	source *telegrafPluginSource
}

func (t *telegrafDataBatch) preparePoints(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	var ts time.Time
	if len(timestamp) > 0 {
		ts = timestamp[0]
	} else {
		ts = time.Now()
	}

	for metric, v := range fields {
		var value float64
		var err error
		switch p := v.(type) {
		case string:
			continue
		case bool:
			if p {
				value = 1
			} else {
				value = 0
			}
		default:
			value, err = getFloat(v)
			if err != nil {
				log.Errorf("unsupported type: %v plugin: %s metric: %v value: %v. error: %v", reflect.TypeOf(v), t.source.name, metric, v, err)
				continue
			}
		}

		metricName := measurement + "." + metric
		metricName = strings.Replace(metricName, "_", ".", -1)
		if len(t.source.prefix) > 0 {
			metricName = t.source.prefix + "." + metricName
		}

		t.Points = wf.FilterAppend(t.source.filters, t.source.pointsFiltered, t.Points, wf.NewPoint(
			metricName,
			value,
			ts.UnixNano()/1000,
			t.source.source,
			t.buildTags(tags),
		))
	}
}

func (t *telegrafDataBatch) buildTags(pointTags map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range t.source.tags {
		if len(v) > 0 {
			result[k] = v
		}
	}
	for k, v := range pointTags {
		if len(v) > 0 {
			result[k] = v
		}
	}
	return result
}

// AddFields adds a metric to the accumulator with the given measurement
// name, fields, and tags (and timestamp). If a timestamp is not provided,
// then the accumulator sets it to "now".
func (t *telegrafDataBatch) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddGauge is the same as AddFields, but will add the metric as a "Gauge" type
func (t *telegrafDataBatch) AddGauge(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddCounter is the same as AddFields, but will add the metric as a "Counter" type
func (t *telegrafDataBatch) AddCounter(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddSummary is the same as AddFields, but will add the metric as a "Summary" type
func (t *telegrafDataBatch) AddSummary(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	log.Fatal("not supported")
}

// AddHistogram is the same as AddFields, but will add the metric as a "Histogram" type
func (t *telegrafDataBatch) AddHistogram(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	log.Fatal("not supported")
}

// AddMetric adds an metric to the accumulator.
func (t *telegrafDataBatch) AddMetric(telegraf.Metric) {
	log.Fatal("not supported")
}

// SetPrecision sets the timestamp rounding precision.  All metrics addeds
// added to the accumulator will have their timestamp rounded to the
// nearest multiple of precision.
func (t *telegrafDataBatch) SetPrecision(precision time.Duration) {
	log.Fatal("not supported")
}

// Report an error.
func (t *telegrafDataBatch) AddError(err error) {
	if err != nil {
		t.source.errors.Inc(1)
		if t.source.targetEPS != nil {
			t.source.targetEPS.Inc(1)
		}
		log.Error(err)
	}
}

// Upgrade to a TrackingAccumulator with space for maxTracked metrics/batches.
func (t *telegrafDataBatch) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	log.Fatal("not supported")
	return nil
}

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (f float64, e error) {
	v := reflect.ValueOf(unk)
	if unk == nil {
		return 0, fmt.Errorf("cannot convert nil value to float64")
	}

	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}
