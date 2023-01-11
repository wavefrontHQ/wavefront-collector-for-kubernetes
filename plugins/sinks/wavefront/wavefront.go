// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package wavefront

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/events"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-sdk-go/senders"

	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

const (
	proxyClient  = 1
	directClient = 2
	testClient   = 3
)

const maxWavefrontTags = 19 // the maximum numbers of tags allowed in a wavefront point not including source

var (
	excludeTagList = [...]string{"namespace_id", "host_id", "pod_id", "hostname"}
	sentPoints     gm.Counter
	errPoints      gm.Counter
	filteredPoints gm.Counter
	sentEvents     gm.Counter
	errEvents      gm.Counter
	clientType     gm.Gauge
	sanitizedChars = strings.NewReplacer("+", "-")
)

func init() {
	sentPoints = gm.GetOrRegisterCounter("wavefront.points.sent.count", gm.DefaultRegistry)
	errPoints = gm.GetOrRegisterCounter("wavefront.points.errors.count", gm.DefaultRegistry)
	filteredPoints = gm.GetOrRegisterCounter("wavefront.points.filtered.count", gm.DefaultRegistry)
	sentEvents = gm.GetOrRegisterCounter("wavefront.events.sent.count", gm.DefaultRegistry)
	errEvents = gm.GetOrRegisterCounter("wavefront.events.errors.count", gm.DefaultRegistry)
	clientType = gm.GetOrRegisterGauge("wavefront.sender.type", gm.DefaultRegistry)
}

type WavefrontSink interface {
	Name() string
	Stop()
	metrics.Sink
	events.EventSink
}

type wavefrontSink struct {
	WavefrontClient senders.Sender
	ClusterName     string
	Prefix          string
	globalTags      map[string]string
	filters         filter.Filter
	forceGC         bool
	logPercent      float32
	stopHeartbeat   chan struct{}
}

func (sink *wavefrontSink) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	name = sanitizedChars.Replace(name)
	if len(sink.Prefix) > 0 {
		name = sink.Prefix + "." + name
	}
	logTagCleaningReasons(name, cleanTags(tags, maxWavefrontTags))
	return sink.WavefrontClient.SendDistribution(name, centroids, hgs, ts, source, tags)
}

func NewWavefrontSink(cfg configuration.WavefrontSinkConfig) (WavefrontSink, error) {
	storage := &wavefrontSink{
		ClusterName: configuration.GetStringValue(cfg.ClusterName, "k8s-cluster"),
		logPercent:  0.01,
	}

	if cfg.EventsEnabled && cfg.ExternalEventsEndpoint != "" {
		//set events only to the externalEventsEndpoint (no event metrics will be reported to wavefront).
		log.Info("Sending Events to ExternalEventsEndpoint:" + cfg.ExternalEventsEndpoint)
		util.SetExternalEventEndpoint(cfg.ExternalEventsEndpoint)
		util.SetToken(cfg.Token)
		storage.WavefrontClient, _ = senders.NewWavefrontNoOpClient()
	} else if cfg.TestMode {
		storage.WavefrontClient = NewTestSender()
		clientType.Update(testClient)
	} else if cfg.ProxyAddress != "" {
		s := strings.Split(cfg.ProxyAddress, ":")
		host, portStr := s[0], s[1]
		port, err := strconv.Atoi(portStr)

		if err != nil {
			return nil, fmt.Errorf("error parsing proxy port: %s", err.Error())
		}
		storage.WavefrontClient, err = senders.NewProxySender(&senders.ProxyConfiguration{
			Host:             host,
			MetricsPort:      port,
			DistributionPort: port,
			EventsPort:       port,
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
			Server:        cfg.Server,
			Token:         cfg.Token,
			BatchSize:     cfg.BatchSize,
			MaxBufferSize: cfg.MaxBufferSize,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating direct sender: %s", err.Error())
		}
		clientType.Update(directClient)
	}
	if storage.WavefrontClient == nil {
		return nil, fmt.Errorf("proxyAddress or server property or externalEventsEndpoint required for sinks")
	}

	storage.globalTags = cfg.Tags
	if cfg.Prefix != "" {
		storage.Prefix = strings.Trim(cfg.Prefix, ".")
	}
	storage.filters = filter.FromConfig(cfg.Filters)

	// force garbage collection if experimental flag enabled
	storage.forceGC = os.Getenv(util.ForceGC) != ""

	// configure error logging percentage
	if cfg.ErrorLogPercent > 0.0 && cfg.ErrorLogPercent <= 1.0 {
		storage.logPercent = cfg.ErrorLogPercent
	}

	// emit heartbeat metric
	storage.emitHeartbeat(storage.WavefrontClient, cfg)

	return storage, nil
}

func (sink *wavefrontSink) Name() string {
	return "wavefront_sink"
}

func (sink *wavefrontSink) Stop() {
	close(sink.stopHeartbeat)
	sink.WavefrontClient.Close()
}

func (sink *wavefrontSink) SendMetric(metricName string, value float64, timestamp int64, source string, tags map[string]string) error {
	metricName = sanitizedChars.Replace(metricName)
	if len(sink.Prefix) > 0 {
		metricName = sink.Prefix + "." + metricName
	}
	logTagCleaningReasons(metricName, cleanTags(tags, maxWavefrontTags))

	return sink.WavefrontClient.SendMetric(metricName, value, timestamp, source, tags)
}

func (sink *wavefrontSink) logVerboseError(f log.Fields, msg string) {
	if log.IsLevelEnabled(log.DebugLevel) {
		log.WithFields(f).Error(msg)
	} else if sink.loggingAllowed() {
		log.WithFields(f).Errorf("%s %s", "[sampled error]", msg)
	}
}

func (sink *wavefrontSink) Export(batch *metrics.Batch) {
	log.Debugf("received metric points: %d", len(batch.Metrics))

	before := errPoints.Count()
	for _, point := range batch.Metrics {
		if point == nil {
			continue
		}
		point.OverrideTag(metrics.LabelCluster.Key, sink.ClusterName)
		point.AddTags(sink.globalTags)
		point = wf.Filter(sink.filters, filteredPoints, point)
		if point == nil {
			continue
		}
		err := point.Send(sink)
		if err != nil {
			errPoints.Inc(1)
			sink.logVerboseError(log.Fields{
				"name":  point.Name(),
				"error": err,
			}, "error sending metric")
		} else {
			sentPoints.Inc(1)
		}
	}

	after := errPoints.Count()
	if after > before {
		log.WithField("count", after).Warning("Error sending one or more points")
	}

	// This seems like an odd place for this considering that we still have references to the big
	// memory user, the Batch. However, moving it until that reference was released actually
	// reduced the effectiveness of this flag. The garbage collector has some interesting ideas about
	// what to clean and when.
	if sink.forceGC {
		log.Info("sink: forcing memory release")
		debug.FreeOSMemory()
	}
}

func (sink *wavefrontSink) ExportEvent(ev *events.Event) {
	ev.Options = append(ev.Options, event.Annotate("cluster", sink.ClusterName))
	host := sink.ClusterName

	if util.GetExternalEventEndpoint() != "" {
		line, _ := senders.EventLineJSON(
			ev.Message,
			ev.Ts.Unix(), 0,
			host,
			ev.Tags,
			ev.Options...,
			)

		req, _ := http.NewRequest("POST", util.GetExternalEventEndpoint(), bytes.NewBuffer([]byte(line)))
		req.Header.Set("Content-Type", "text/plain")
		fmt.Printf("TEST:: token: " + util.GetToken())
		req.Header.Set("Authorization", "Bearer " + util.GetToken())

		client := &http.Client{}
		_, err := client.Do(req)
		if err != nil {
			sink.logVerboseError(log.Fields{
				"message": ev.Message,
				"error":   err,
			}, "error sending event to external event endpoint")
		}
	} else {
		err := sink.WavefrontClient.SendEvent(
			ev.Message,
			ev.Ts.Unix(), 0,
			host,
			ev.Tags,
			ev.Options...,
		)
		if err != nil {
			sink.logVerboseError(log.Fields{
				"message": ev.Message,
				"error":   err,
			}, "error sending event")
			errEvents.Inc(1)
		} else {
			sentEvents.Inc(1)
		}
	}
}

func getDefault(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func (sink *wavefrontSink) loggingAllowed() bool {
	return rand.Float32() <= sink.logPercent
}

func (sink *wavefrontSink) emitHeartbeat(sender senders.Sender, cfg configuration.WavefrontSinkConfig) {
	ticker := time.NewTicker(1 * time.Minute)
	sink.stopHeartbeat = make(chan struct{})
	source := getDefault(util.GetNodeName(), "wavefront-collector-for-kubernetes")
	tags := map[string]string{
		"cluster":             cfg.ClusterName,
		"stats_prefix":        configuration.GetStringValue(cfg.Prefix, "kubernetes."),
		"installation_method": util.GetInstallationMethod(),
	}

	eventsEnabled := 0.0
	if cfg.EventsEnabled {
		eventsEnabled = 1.0
	}

	go func() {
		log.Debug("emitting heartbeat metric")
		util.AddK8sTags(tags)
		err := sender.SendMetric("~wavefront.kubernetes.collector.version", cfg.Version, 0, source, tags)
		if err != nil {
			log.Debugf("error emitting heartbeat metric :%v", err)
		}
		for {
			select {
			case <-ticker.C:
				util.AddK8sTags(tags)
				_ = sender.SendMetric("~wavefront.kubernetes.collector.version", cfg.Version, 0, source, tags)
				_ = sender.SendMetric("~wavefront.kubernetes.collector.config.events.enabled", eventsEnabled, 0, source, tags)
				sink.logStatus()
			case <-sink.stopHeartbeat:
				log.Info("stopping heartbeat")
				ticker.Stop()
				return
			}
		}
	}()
}

func (sink *wavefrontSink) logStatus() {
	// # events can be large in volume. log a status message periodically
	sent := sentEvents.Count()
	errs := errEvents.Count()
	if sent > 0 || errs > 0 {
		log.WithFields(log.Fields{
			"sent":   sentEvents.Count(),
			"errors": errEvents.Count(),
		}).Info("Events processed")
	}
}
