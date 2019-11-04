// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package wavefront

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/events"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-sdk-go/senders"

	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

const (
	proxyClient  = 1
	directClient = 2
)

var (
	excludeTagList = [...]string{"namespace_id", "host_id", "pod_id", "hostname"}
	sentPoints     gm.Counter
	errPoints      gm.Counter
	msCount        gm.Counter
	filteredPoints gm.Counter
	clientType     gm.Gauge
	sanitizedChars = strings.NewReplacer("+", "-")
)

func init() {
	sentPoints = gm.GetOrRegisterCounter("wavefront.points.sent.count", gm.DefaultRegistry)
	errPoints = gm.GetOrRegisterCounter("wavefront.points.errors.count", gm.DefaultRegistry)
	msCount = gm.GetOrRegisterCounter("wavefront.points.metric-sets.count", gm.DefaultRegistry)
	filteredPoints = gm.GetOrRegisterCounter("wavefront.points.filtered.count", gm.DefaultRegistry)
	clientType = gm.GetOrRegisterGauge("wavefront.sender.type", gm.DefaultRegistry)
}

type WavefrontSink interface {
	Name() string
	Stop()
	metrics.DataSink
	events.EventSink
}

type wavefrontSink struct {
	WavefrontClient   senders.Sender
	ClusterName       string
	Prefix            string
	globalTags        map[string]string
	filters           filter.Filter
	testMode          bool
	testReceivedLines []string
}

func (sink *wavefrontSink) Name() string {
	return "wavefront_sink"
}

func (sink *wavefrontSink) Stop() {
	sink.WavefrontClient.Close()
}

func (sink *wavefrontSink) sendPoint(metricName string, value float64, ts int64, source string, tags map[string]string) {
	metricName = sanitizedChars.Replace(metricName)
	if sink.filters != nil && !sink.filters.Match(metricName, tags) {
		filteredPoints.Inc(1)
		log.WithField("name", metricName).Trace("Dropping metric")
		return
	}

	tags = combineGlobalTags(tags, sink.globalTags)

	if sink.testMode {
		tagStr := ""
		for k, v := range tags {
			tagStr += k + "=\"" + v + "\" "
		}
		line := fmt.Sprintf("%s %f %d source=\"%s\" %s\n", metricName, value, ts, source, tagStr)
		sink.testReceivedLines = append(sink.testReceivedLines, line)
		log.Infoln(line)
		return
	}
	err := sink.WavefrontClient.SendMetric(metricName, value, ts, source, tags)
	if err != nil {
		errPoints.Inc(1)
		log.WithFields(log.Fields{
			"name":  metricName,
			"error": err,
		}).Debug("error sending metric")
	} else {
		sentPoints.Inc(1)
	}
}

func combineGlobalTags(tags, globalTags map[string]string) map[string]string {
	if tags == nil || len(tags) == 0 {
		return globalTags
	}
	if globalTags == nil || len(globalTags) == 0 {
		return tags
	}

	for k, v := range globalTags {
		// add global tag if key is missing from tags
		if _, exists := tags[k]; !exists {
			tags[k] = v
		}
	}
	return tags
}

func (sink *wavefrontSink) send(batch *metrics.DataBatch) {
	log.Debugf("received metric points: %d", len(batch.MetricPoints))

	before := errPoints.Count()
	for _, point := range batch.MetricPoints {
		tags := make(map[string]string)

		for k, v := range point.Tags {
			if len(v) > 0 {
				tags[k] = v
			}
		}

		if len(point.StrTags) > 0 {
			for _, tag := range strings.Split(point.StrTags, " ") {
				if len(tag) > 0 {
					s := strings.Split(tag, "=")
					if len(s) == 2 {
						k, v := s[0], s[1]
						if len(v) > 0 {
							key, _ := url.QueryUnescape(k)
							val, _ := url.QueryUnescape(v)
							if len(key) > 0 && len(val) > 0 {
								tags[key] = val
							}
						}
					}
				}
			}
		}
		tags["cluster"] = sink.ClusterName
		sink.sendPoint(point.Metric, point.Value, point.Timestamp, point.Source, tags)
	}

	after := errPoints.Count()
	if after > before {
		log.WithField("count", after).Warning("Error sending one or more points")
	}
}

func (sink *wavefrontSink) ExportData(batch *metrics.DataBatch) {
	if sink.testMode {
		//clear lines from last batch
		sink.testReceivedLines = sink.testReceivedLines[:0]
		sink.send(batch)
		return
	}
	sink.send(batch)
}

func (wf *wavefrontSink) ExportEvent(event *events.Event) {
	event.Tags["cluster"] = wf.ClusterName

	err := wf.WavefrontClient.SendEvent(
		event.Message,
		event.Ts.Unix(), 0,
		event.Host,
		event.Tags,
		event.Options...,
	)
	if err != nil {
		log.Error(err)
	}
}

func NewWavefrontSink(cfg configuration.WavefrontSinkConfig) (WavefrontSink, error) {
	storage := &wavefrontSink{
		ClusterName: configuration.GetStringValue(cfg.ClusterName, "k8s-cluster"),
		testMode:    cfg.TestMode,
	}

	if cfg.ProxyAddress != "" {
		s := strings.Split(cfg.ProxyAddress, ":")
		host, portStr := s[0], s[1]
		port, err := strconv.Atoi(portStr)

		if err != nil {
			return nil, fmt.Errorf("error parsing proxy port: %s", err.Error())
		}
		storage.WavefrontClient, err = senders.NewProxySender(&senders.ProxyConfiguration{
			Host:        host,
			MetricsPort: port,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating proxy sender: %s", err.Error())
		}
		clientType.Update(proxyClient)
	} else if cfg.Server != "" {
		if len(cfg.Token) == 0 {
			return nil, fmt.Errorf("token missing for Wavefront sink")
		}
		var err error
		storage.WavefrontClient, err = senders.NewDirectSender(&senders.DirectConfiguration{
			Server: cfg.Server,
			Token:  cfg.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating direct sender: %s", err.Error())
		}
		clientType.Update(directClient)
	}
	if storage.WavefrontClient == nil {
		return nil, fmt.Errorf("proxyAddress or server property required for Wavefront sink")
	}

	storage.globalTags = cfg.Tags
	if cfg.Prefix != "" {
		storage.Prefix = cfg.Prefix
	}
	storage.filters = filter.FromConfig(cfg.Filters)

	return storage, nil
}
