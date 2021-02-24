package senders

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
)

type directSender struct {
	reporter         internal.Reporter
	defaultSource    string
	pointHandler     *internal.LineHandler
	histoHandler     *internal.LineHandler
	spanHandler      *internal.LineHandler
	spanLogHandler   *internal.LineHandler
	eventHandler     *internal.LineHandler
	internalRegistry *internal.MetricRegistry
}

// NewDirectSender creates and returns a Wavefront Direct Ingestion Sender instance
func NewDirectSender(cfg *DirectConfiguration) (Sender, error) {
	if cfg.Server == "" || cfg.Token == "" {
		return nil, fmt.Errorf("server and token cannot be empty")
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = defaultBatchSize
	}
	if cfg.MaxBufferSize == 0 {
		cfg.MaxBufferSize = defaultBufferSize
	}
	if cfg.FlushIntervalSeconds == 0 {
		cfg.FlushIntervalSeconds = defaultFlushInterval
	}

	reporter := internal.NewDirectReporter(cfg.Server, cfg.Token)

	sender := &directSender{
		defaultSource: internal.GetHostname("wavefront_direct_sender"),
	}
	sender.internalRegistry = internal.NewMetricRegistry(
		sender,
		internal.SetPrefix("~sdk.go.core.sender.direct"),
		internal.SetTag("pid", strconv.Itoa(os.Getpid())),
	)
	sender.pointHandler = makeLineHandler(reporter, cfg, internal.MetricFormat, "points", sender.internalRegistry)
	sender.histoHandler = makeLineHandler(reporter, cfg, internal.HistogramFormat, "histograms", sender.internalRegistry)
	sender.spanHandler = makeLineHandler(reporter, cfg, internal.TraceFormat, "spans", sender.internalRegistry)
	sender.spanLogHandler = makeLineHandler(reporter, cfg, internal.SpanLogsFormat, "span_logs", sender.internalRegistry)
	sender.eventHandler = makeLineHandler(reporter, cfg, internal.EventFormat, "events", sender.internalRegistry)

	sender.Start()
	return sender, nil
}

func makeLineHandler(reporter internal.Reporter, cfg *DirectConfiguration, format, prefix string,
	registry *internal.MetricRegistry) *internal.LineHandler {
	flushInterval := time.Second * time.Duration(cfg.FlushIntervalSeconds)

	opts := []internal.LineHandlerOption{internal.SetHandlerPrefix(prefix), internal.SetRegistry(registry)}
	batchSize := cfg.BatchSize
	if format == internal.EventFormat {
		batchSize = 1
		opts = append(opts, internal.SetLockOnThrottledError(true))
	}

	return internal.NewLineHandler(reporter, format, flushInterval, batchSize, cfg.MaxBufferSize, opts...)
}

func (sender *directSender) Start() {
	sender.pointHandler.Start()
	sender.histoHandler.Start()
	sender.spanHandler.Start()
	sender.spanLogHandler.Start()
	sender.internalRegistry.Start()
	sender.eventHandler.Start()
}

func (sender *directSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	line, err := MetricLine(name, value, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	return sender.pointHandler.HandleLine(line)
}

func (sender *directSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if name == "" {
		return fmt.Errorf("empty metric name")
	}
	if !internal.HasDeltaPrefix(name) {
		name = internal.DeltaCounterName(name)
	}
	if value > 0 {
		return sender.SendMetric(name, value, 0, source, tags)
	}
	return nil
}

func (sender *directSender) SendDistribution(name string, centroids []histogram.Centroid,
	hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	line, err := HistoLine(name, centroids, hgs, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	return sender.histoHandler.HandleLine(line)
}

func (sender *directSender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string,
	parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	line, err := SpanLine(name, startMillis, durationMillis, source, traceId, spanId, parents, followsFrom, tags, spanLogs, sender.defaultSource)
	if err != nil {
		return err
	}
	err = sender.spanHandler.HandleLine(line)
	if err != nil {
		return err
	}

	if len(spanLogs) > 0 {
		logs, err := SpanLogJSON(traceId, spanId, spanLogs)
		if err != nil {
			return err
		}
		return sender.spanLogHandler.HandleLine(logs)
	}
	return nil
}

func (sender *directSender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	line, err := EventLineJSON(name, startMillis, endMillis, source, tags, setters...)
	if err != nil {
		return err
	}
	return sender.eventHandler.HandleLine(line)
}

func (sender *directSender) Close() {
	sender.pointHandler.Stop()
	sender.histoHandler.Stop()
	sender.spanHandler.Stop()
	sender.spanLogHandler.Stop()
	sender.internalRegistry.Stop()
	sender.eventHandler.Stop()
}

func (sender *directSender) Flush() error {
	errStr := ""
	err := sender.pointHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = sender.histoHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = sender.spanHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	err = sender.spanLogHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	err = sender.eventHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

func (sender *directSender) GetFailureCount() int64 {
	return sender.pointHandler.GetFailureCount() +
		sender.histoHandler.GetFailureCount() +
		sender.spanHandler.GetFailureCount() +
		sender.spanLogHandler.GetFailureCount() +
		sender.eventHandler.GetFailureCount()
}
