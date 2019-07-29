// Based on https://github.com/prometheus/node_exporter/blob/master/collector/systemd_linux.go.
// Diff against commit f028b816152f6d5650ca2cd707e45cda7333fdc1so for changes to the original code.

// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package systemd

import (
	"fmt"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	"github.com/coreos/go-systemd/dbus"
	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

var unitStatesName = []string{"active", "activating", "deactivating", "inactive", "failed"}

type systemdMetricsSource struct {
	prefix                  string
	source                  string
	collectTaskMetrics      bool
	collectStartTimeMetrics bool
	collectRestartMetrics   bool
	unitsFilter             *unitFilter
	filters                 filter.Filter

	pps gm.Counter
	fps gm.Counter
	eps gm.Counter
}

func (src *systemdMetricsSource) Name() string {
	return "systemd_metrics_source"
}

func (src *systemdMetricsSource) ScrapeMetrics() (*DataBatch, error) {
	// gathers metrics from systemd using dbus. collection is done in parallel to reduce wait time for responses.
	conn, err := dbus.New()
	if err != nil {
		src.eps.Inc(1)
		return nil, fmt.Errorf("couldn't get dbus connection: %s", err)
	}
	defer conn.Close()

	allUnits, err := src.getAllUnits(conn)
	if err != nil {
		src.eps.Inc(1)
		return nil, fmt.Errorf("couldn't get units: %s", err)
	}

	now := time.Now().Unix()
	result := &DataBatch{
		Timestamp: time.Now(),
	}

	// channel for gathering collected metrics
	gather := make(chan *MetricPoint, 1000)
	done := make(chan bool)
	var points []*MetricPoint

	// goroutine for gathering collected metrics
	go func() {
		for {
			select {
			case point, more := <-gather:
				if !more {
					log.Infof("systemd metrics collection complete")
					done <- true
					return
				}
				points = src.filterAppend(points, point)
			}
		}
	}()

	summary := summarizeUnits(allUnits)
	src.collectSummaryMetrics(summary, gather, now)

	units := src.filterUnits(allUnits)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		src.collectUnitStatusMetrics(conn, units, gather, now)
	}()

	if src.collectStartTimeMetrics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			src.collectUnitStartTimeMetrics(conn, units, gather, now)
		}()
	}

	if src.collectTaskMetrics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			src.collectUnitTasksMetrics(conn, units, gather, now)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		src.collectTimers(conn, units, gather, now)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		src.collectSockets(conn, units, gather, now)
	}()

	err = src.collectSystemState(conn, gather, now)
	if err != nil {
		src.eps.Inc(1)
		log.Errorf("error collecting system stats: %v", err)
	}

	// wait for collection to complete and then close the gathering channel
	wg.Wait()
	close(gather)

	// wait for gathering to process all the points
	<-done

	result.MetricPoints = points
	count := len(result.MetricPoints)
	log.Infof("%s metrics: %d", "systemd", count)
	src.pps.Inc(int64(count))

	return result, err
}

func (src *systemdMetricsSource) collectUnitStatusMetrics(conn *dbus.Conn, units []unit, ch chan<- *MetricPoint, now int64) {
	for _, unit := range units {
		serviceType := ""
		if strings.HasSuffix(unit.Name, ".service") {
			serviceTypeProperty, err := conn.GetUnitTypeProperty(unit.Name, "Service", "Type")
			if err != nil {
				log.Infof("couldn't get unit '%s' Type: %s", unit.Name, err)
			} else {
				serviceType = serviceTypeProperty.Value.Value().(string)
			}
		} else if strings.HasSuffix(unit.Name, ".mount") {
			serviceTypeProperty, err := conn.GetUnitTypeProperty(unit.Name, "Mount", "Type")
			if err != nil {
				log.Debugf("couldn't get unit '%s' Type: %s", unit.Name, err)
			} else {
				serviceType = serviceTypeProperty.Value.Value().(string)
			}
		}
		for _, stateName := range unitStatesName {
			isActive := 0.0
			if stateName == unit.ActiveState {
				isActive = 1.0
			}
			tags := map[string]string{}
			setTags(tags, unit.Name, stateName, serviceType)
			ch <- src.metricPoint("unit_state", isActive, now, tags)
		}
		if src.collectRestartMetrics && strings.HasSuffix(unit.Name, ".service") {
			// NRestarts wasn't added until systemd 235.
			restartsCount, err := conn.GetUnitTypeProperty(unit.Name, "Service", "NRestarts")
			if err != nil {
				log.Debugf("couldn't get unit '%s' NRestarts: %s", unit.Name, err)
			} else {
				tags := map[string]string{}
				setTag(tags, "name", unit.Name)
				ch <- src.metricPoint("service_restart_total", float64(restartsCount.Value.Value().(uint32)), now, tags)
			}
		}
	}
}

func (src *systemdMetricsSource) collectSockets(conn *dbus.Conn, units []unit, ch chan<- *MetricPoint, now int64) {
	for _, unit := range units {
		if !strings.HasSuffix(unit.Name, ".socket") {
			continue
		}

		acceptedConnectionCount, err := conn.GetUnitTypeProperty(unit.Name, "Socket", "NAccepted")
		if err != nil {
			log.Debugf("couldn't get unit '%s' NAccepted: %s", unit.Name, err)
			continue
		}
		tags := map[string]string{}
		setTag(tags, "name", unit.Name)
		ch <- src.metricPoint("socket_accepted_connections_total", float64(acceptedConnectionCount.Value.Value().(uint32)), now, tags)

		currentConnectionCount, err := conn.GetUnitTypeProperty(unit.Name, "Socket", "NConnections")
		if err != nil {
			log.Debugf("couldn't get unit '%s' NConnections: %s", unit.Name, err)
			continue
		}
		ch <- src.metricPoint("socket_current_connections", float64(currentConnectionCount.Value.Value().(uint32)), now, tags)

		// NRefused wasn't added until systemd 239.
		refusedConnectionCount, err := conn.GetUnitTypeProperty(unit.Name, "Socket", "NRefused")
		if err != nil {
			log.Debugf("couldn't get unit '%s' NRefused: %s", unit.Name, err)
		} else {
			ch <- src.metricPoint("socket_refused_connections_total", float64(refusedConnectionCount.Value.Value().(uint32)), now, tags)
		}
	}
}

func (src *systemdMetricsSource) collectUnitStartTimeMetrics(conn *dbus.Conn, units []unit, ch chan<- *MetricPoint, now int64) {
	var startTimeUsec uint64
	for _, unit := range units {
		if unit.ActiveState != "active" {
			startTimeUsec = 0
		} else {
			timestampValue, err := conn.GetUnitProperty(unit.Name, "ActiveEnterTimestamp")
			if err != nil {
				log.Debugf("couldn't get unit '%s' StartTimeUsec: %s", unit.Name, err)
				continue
			}
			startTimeUsec = timestampValue.Value.Value().(uint64)
		}
		tags := map[string]string{}
		setTag(tags, "name", unit.Name)
		ch <- src.metricPoint("unit_start_time_seconds", float64(startTimeUsec)/1e6, now, tags)
	}
}

func (src *systemdMetricsSource) collectUnitTasksMetrics(conn *dbus.Conn, units []unit, ch chan<- *MetricPoint, now int64) {
	var val uint64
	for _, unit := range units {
		if strings.HasSuffix(unit.Name, ".service") {
			tasksCurrentCount, err := conn.GetUnitTypeProperty(unit.Name, "Service", "TasksCurrent")
			if err != nil {
				log.Infof("couldn't get unit '%s' TasksCurrent: %s", unit.Name, err)
			} else {
				val = tasksCurrentCount.Value.Value().(uint64)
				// Don't set if tasksCurrent if dbus reports MaxUint64.
				if val != math.MaxUint64 {
					tags := map[string]string{}
					setTag(tags, "name", unit.Name)
					ch <- src.metricPoint("unit_tasks_current", float64(val), now, tags)
				}
			}
			tasksMaxCount, err := conn.GetUnitTypeProperty(unit.Name, "Service", "TasksMax")
			if err != nil {
				log.Infof("couldn't get unit '%s' TasksMax: %s", unit.Name, err)
			} else {
				val = tasksMaxCount.Value.Value().(uint64)
				// Don't set if tasksMax if dbus reports MaxUint64.
				if val != math.MaxUint64 {
					tags := map[string]string{}
					setTag(tags, "name", unit.Name)
					ch <- src.metricPoint("unit_tasks_max", float64(val), now, tags)
				}
			}
		}
	}
}

func (src *systemdMetricsSource) collectTimers(conn *dbus.Conn, units []unit, ch chan<- *MetricPoint, now int64) {
	for _, unit := range units {
		if !strings.HasSuffix(unit.Name, ".timer") {
			continue
		}

		lastTriggerValue, err := conn.GetUnitTypeProperty(unit.Name, "Timer", "LastTriggerUSec")
		if err != nil {
			log.Debugf("couldn't get unit '%s' LastTriggerUSec: %s", unit.Name, err)
			continue
		}
		tags := map[string]string{}
		setTag(tags, "name", unit.Name)
		ch <- src.metricPoint("timer_last_trigger_seconds", float64(lastTriggerValue.Value.Value().(uint64))/1e6, now, tags)
	}
}

func (src *systemdMetricsSource) collectSummaryMetrics(summary map[string]float64, ch chan<- *MetricPoint, now int64) {
	for stateName, count := range summary {
		tags := map[string]string{}
		setTag(tags, "state_name", stateName)
		ch <- src.metricPoint("units", count, now, tags)
	}
}

func (src *systemdMetricsSource) collectSystemState(conn *dbus.Conn, ch chan<- *MetricPoint, now int64) error {
	systemState, err := conn.GetManagerProperty("SystemState")
	if err != nil {
		return fmt.Errorf("couldn't get system state: %s", err)
	}
	isSystemRunning := 0.0
	if systemState == `"running"` {
		isSystemRunning = 1.0
	}
	ch <- src.metricPoint("system_running", isSystemRunning, now, nil)
	return nil
}

type unit struct {
	dbus.UnitStatus
}

func (src *systemdMetricsSource) getAllUnits(conn *dbus.Conn) ([]unit, error) {
	units, err := conn.ListUnits()
	if err != nil {
		return nil, err
	}

	result := make([]unit, 0, len(units))
	for _, status := range units {
		unit := unit{
			UnitStatus: status,
		}
		result = append(result, unit)
	}
	return result, nil
}

func (src *systemdMetricsSource) filterUnits(units []unit) []unit {
	filtered := make([]unit, 0, len(units))
	for _, unit := range units {
		if (src.unitsFilter == nil || src.unitsFilter.match(unit.Name)) && unit.LoadState == "loaded" {
			log.Debugf("Adding unit: %s", unit.Name)
			filtered = append(filtered, unit)
		} else {
			log.Debugf("Ignoring unit: %s", unit.Name)
		}
	}
	return filtered
}

func (src *systemdMetricsSource) filterAppend(slice []*MetricPoint, point *MetricPoint) []*MetricPoint {
	if src.filters == nil || src.filters.Match(point.Metric, point.Tags) {
		return append(slice, point)
	}
	src.fps.Inc(1)
	log.Debugf("dropping metric: %s", point.Metric)
	return slice
}

func summarizeUnits(units []unit) map[string]float64 {
	summarized := make(map[string]float64)
	for _, unitStateName := range unitStatesName {
		summarized[unitStateName] = 0.0
	}
	for _, unit := range units {
		summarized[unit.ActiveState] += 1.0
	}
	return summarized
}

func setTags(tags map[string]string, name, state, service string) {
	setTag(tags, "name", name)
	setTag(tags, "state_name", state)
	setTag(tags, "type", service)
}

func setTag(tags map[string]string, key, val string) {
	if val != "" {
		tags[key] = val
	}
}

func (src *systemdMetricsSource) metricPoint(name string, value float64, ts int64, tags map[string]string) *MetricPoint {
	return &MetricPoint{
		Metric:    src.prefix + strings.Replace(name, "_", ".", -1),
		Value:     value,
		Timestamp: ts,
		Source:    src.source,
		Tags:      tags,
	}
}

type systemdProvider struct {
	metrics.DefaultMetricsSourceProvider
	sources []MetricsSource
}

func (sp *systemdProvider) GetMetricsSources() []MetricsSource {
	return sp.sources
}

func (sp *systemdProvider) Name() string {
	return "systemd_provider"
}

func NewProvider(uri *url.URL) (MetricsSourceProvider, error) {
	vals := uri.Query()

	prefix := "kubernetes.systemd."
	if len(vals["prefix"]) > 0 {
		prefix = vals["prefix"][0]
	}

	source := util.GetNodeName()
	if source == "" {
		var err error
		source, err = os.Hostname()
		if err != nil {
			source = "wavefront-kubernetes-collector"
		}
	}

	collectTaskMetrics := true
	if len(vals["taskMetrics"]) > 0 {
		var err error
		collectTaskMetrics, err = strconv.ParseBool(vals["taskMetrics"][0])
		if err != nil {
			log.Infof("error parsing taskMetrics property: %v", err)
		}
	}

	collectStartTimeMetrics := true
	if len(vals["startTimeMetrics"]) > 0 {
		var err error
		collectStartTimeMetrics, err = strconv.ParseBool(vals["startTimeMetrics"][0])
		if err != nil {
			log.Infof("error parsing startTimeMetrics property: %v", err)
		}
	}

	collectRestartMetrics := false
	if len(vals["restartMetrics"]) > 0 {
		var err error
		collectRestartMetrics, err = strconv.ParseBool(vals["restartMetrics"][0])
		if err != nil {
			log.Infof("error parsing restartMetrics property: %v", err)
		}
	}

	unitsFilter := fromQuery(vals)
	filters := filter.FromQuery(vals)

	pt := map[string]string{"type": "systemd"}
	ppsKey := reporting.EncodeKey("source.points.collected", pt)
	fpsKey := reporting.EncodeKey("source.points.filtered", pt)
	epsKey := reporting.EncodeKey("source.collect.errors", pt)

	sources := make([]MetricsSource, 1)
	sources[0] = &systemdMetricsSource{
		prefix:                  prefix,
		source:                  source,
		collectTaskMetrics:      collectTaskMetrics,
		collectStartTimeMetrics: collectStartTimeMetrics,
		collectRestartMetrics:   collectRestartMetrics,
		unitsFilter:             unitsFilter,
		filters:                 filters,
		pps:                     gm.GetOrRegisterCounter(ppsKey, gm.DefaultRegistry),
		fps:                     gm.GetOrRegisterCounter(fpsKey, gm.DefaultRegistry),
		eps:                     gm.GetOrRegisterCounter(epsKey, gm.DefaultRegistry),
	}

	return &systemdProvider{
		sources: sources,
	}, nil
}
