// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package processors

import (
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	log "github.com/sirupsen/logrus"
)

type RateCalculator struct {
	rateMetricsMapping map[string]metrics.Metric

	lock               sync.Mutex
	previousMetricSets map[metrics.ResourceKey]*metrics.Set
	cachePruneInterval time.Duration
	lastPruneTime      time.Time
}

func (rc *RateCalculator) Name() string {
	return "rate calculator"
}

func (rc *RateCalculator) Process(batch *metrics.Batch) (*metrics.Batch, error) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	for key, newMs := range batch.Sets {
		oldMs, found := rc.previousMetricSets[key]
		if !found {
			log.Debugf("Skipping rates for '%s' - no previous batch found", key)
			rc.previousMetricSets[key] = newMs
			continue
		}

		if !newMs.ScrapeTime.After(oldMs.ScrapeTime) {
			// New must be strictly after old.
			log.Debugf("Skipping rates for '%s' - new batch (%s) was not scraped strictly after old batch (%s)", key, newMs.ScrapeTime, oldMs.ScrapeTime)
			continue
		}
		if !newMs.CollectionStartTime.Equal(oldMs.CollectionStartTime) {
			log.Debugf("Skipping rates for '%s' - different collection start time (restart) new:%v  old:%v", key, newMs.CollectionStartTime, oldMs.CollectionStartTime)
			rc.previousMetricSets[key] = newMs
			continue
		}

		var metricValNew, metricValOld metrics.Value
		var foundNew, foundOld bool

		for metricName, targetMetric := range rc.rateMetricsMapping {
			if metricName == metrics.MetricDiskIORead.MetricDescriptor.Name || metricName == metrics.MetricDiskIOWrite.MetricDescriptor.Name {
				for _, itemNew := range newMs.LabeledValues {
					foundNew, foundOld = false, false
					if itemNew.Name == metricName {
						metricValNew, foundNew = itemNew.Value, true
						for _, itemOld := range oldMs.LabeledValues {
							// Fix negative value on "disk/io_read_bytes_rate" and "disk/io_write_bytes_rate" when multiple disk devices are available
							if itemOld.Name == metricName && itemOld.Labels[metrics.LabelResourceID.Key] == itemNew.Labels[metrics.LabelResourceID.Key] {
								metricValOld, foundOld = itemOld.Value, true
								break
							}
						}
					}

					if foundNew && foundOld {
						if targetMetric.MetricDescriptor.ValueType == metrics.ValueFloat {
							newVal := 1e9 * float64(metricValNew.IntValue-metricValOld.IntValue) /
								float64(newMs.ScrapeTime.UnixNano()-oldMs.ScrapeTime.UnixNano())

							newMs.LabeledValues = append(newMs.LabeledValues, metrics.LabeledValue{
								Name:   targetMetric.MetricDescriptor.Name,
								Labels: itemNew.Labels,
								Value: metrics.Value{
									ValueType:  metrics.ValueFloat,
									FloatValue: newVal,
								},
							})
						}
					} else if foundNew && !foundOld || !foundNew && foundOld {
						log.Debugf("Skipping rates for '%s' in '%s': metric not found in one of old (%v) or new (%v)", metricName, key, foundOld, foundNew)
					}
				}
			} else {
				metricValNew, foundNew = newMs.Values[metricName]
				metricValOld, foundOld = oldMs.Values[metricName]

				if foundNew && foundOld && metricName == metrics.MetricCpuUsage.MetricDescriptor.Name {
					// cpu/usage values are in nanoseconds; we want to have it in millicores (that's why constant 1000 is here).
					newVal := 1000 * (metricValNew.IntValue - metricValOld.IntValue) /
						(newMs.ScrapeTime.UnixNano() - oldMs.ScrapeTime.UnixNano())

					newMs.Values[targetMetric.MetricDescriptor.Name] = metrics.Value{
						ValueType: metrics.ValueInt64,
						IntValue:  newVal,
					}

				} else if foundNew && foundOld && targetMetric.MetricDescriptor.ValueType == metrics.ValueFloat {
					newVal := 1e9 * float64(metricValNew.IntValue-metricValOld.IntValue) /
						float64(newMs.ScrapeTime.UnixNano()-oldMs.ScrapeTime.UnixNano())

					newMs.Values[targetMetric.MetricDescriptor.Name] = metrics.Value{
						ValueType:  metrics.ValueFloat,
						FloatValue: newVal,
					}
				} else if foundNew && !foundOld || !foundNew && foundOld {
					log.Debugf("Skipping rates for '%s' in '%s': metric not found in one of old (%v) or new (%v)", metricName, key, foundOld, foundNew)
				}
			}
		}
		rc.previousMetricSets[key] = newMs
	}

	// periodically prune deleted pods, containers etc from the internal cache
	log.Debugf("rate cache size: %d", len(rc.previousMetricSets))
	if rc.lastPruneTime.Before(time.Now().Add(-1 * rc.cachePruneInterval)) {
		log.Infof("pruning rate cache. cache size: %d lastPruneTime: %v", len(rc.previousMetricSets), rc.lastPruneTime)
		for key := range rc.previousMetricSets {
			if _, found := batch.Sets[key]; !found {
				log.Debugf("removing key %s from rate cache", key)
				delete(rc.previousMetricSets, key)
			}
		}
		rc.lastPruneTime = time.Now()
		log.Infof("cache pruning completed. cache size: %d", len(rc.previousMetricSets))
	}
	return batch, nil
}

func NewRateCalculator(rateMetricsMapping map[string]metrics.Metric) *RateCalculator {
	return &RateCalculator{
		rateMetricsMapping: rateMetricsMapping,
		previousMetricSets: make(map[metrics.ResourceKey]*metrics.Set, 256),
		cachePruneInterval: 5 * time.Minute,
		lastPruneTime:      time.Now(),
	}
}
