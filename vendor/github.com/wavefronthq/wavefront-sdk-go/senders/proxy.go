package senders

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
)

const (
	metricHandler int = iota
	histoHandler
	spanHandler
	eventHandler
	handlersCount
)

type proxySender struct {
	handlers      []internal.ConnectionHandler
	defaultSource string
}

// Creates and returns a Wavefront Proxy Sender instance
func NewProxySender(cfg *ProxyConfiguration) (Sender, error) {
	sender := &proxySender{
		defaultSource: internal.GetHostname("wavefront_proxy_sender"),
		handlers:      make([]internal.ConnectionHandler, handlersCount),
	}

	if cfg.FlushIntervalSeconds == 0 {
		cfg.FlushIntervalSeconds = defaultProxyFlushInterval
	}

	if cfg.MetricsPort != 0 {
		sender.handlers[metricHandler] = makeConnHandler(cfg.Host, cfg.MetricsPort, cfg.FlushIntervalSeconds)
	}

	if cfg.DistributionPort != 0 {
		sender.handlers[histoHandler] = makeConnHandler(cfg.Host, cfg.DistributionPort, cfg.FlushIntervalSeconds)
	}

	if cfg.TracingPort != 0 {
		sender.handlers[spanHandler] = makeConnHandler(cfg.Host, cfg.TracingPort, cfg.FlushIntervalSeconds)
	}

	if cfg.EventsPort != 0 {
		sender.handlers[eventHandler] = makeConnHandler(cfg.Host, cfg.EventsPort, cfg.FlushIntervalSeconds)
	}

	for _, h := range sender.handlers {
		if h != nil {
			sender.Start()
			return sender, nil
		}
	}

	return nil, errors.New("at least one proxy port should be enabled")
}

func makeConnHandler(host string, port, flushIntervalSeconds int) internal.ConnectionHandler {
	addr := host + ":" + strconv.FormatInt(int64(port), 10)
	flushInterval := time.Second * time.Duration(flushIntervalSeconds)
	return internal.NewProxyConnectionHandler(addr, flushInterval)
}

func (sender *proxySender) Start() {
	for _, h := range sender.handlers {
		if h != nil {
			h.Start()
		}
	}
}

func (sender *proxySender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	handler := sender.handlers[metricHandler]
	if handler == nil {
		return errors.New("proxy metrics port not provided, cannot send metric data")
	}

	if !handler.Connected() {
		if err := handler.Connect(); err != nil {
			return err
		}
	}

	line, err := MetricLine(name, value, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	err = handler.SendData(line)
	return err
}

func (sender *proxySender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if name == "" {
		return errors.New("empty metric name")
	}
	if !internal.HasDeltaPrefix(name) {
		name = internal.DeltaCounterName(name)
	}
	if value > 0 {
		return sender.SendMetric(name, value, 0, source, tags)
	}
	return nil
}

func (sender *proxySender) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	handler := sender.handlers[histoHandler]
	if handler == nil {
		return errors.New("proxy distribution port not provided, cannot send distribution data")
	}

	if !handler.Connected() {
		if err := handler.Connect(); err != nil {
			return err
		}
	}

	line, err := HistoLine(name, centroids, hgs, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	err = handler.SendData(line)
	return err
}

func (sender *proxySender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string, parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	handler := sender.handlers[spanHandler]
	if handler == nil {
		return errors.New("proxy tracing port not provided, cannot send span data")
	}

	if !handler.Connected() {
		if err := handler.Connect(); err != nil {
			return err
		}
	}

	line, err := SpanLine(name, startMillis, durationMillis, source, traceId, spanId, parents, followsFrom, tags, spanLogs, sender.defaultSource)
	if err != nil {
		return err
	}
	err = handler.SendData(line)
	if err != nil {
		return err
	}

	if len(spanLogs) > 0 {
		logs, err := SpanLogJSON(traceId, spanId, spanLogs)
		if err != nil {
			return err
		}
		return handler.SendData(logs)
	}
	return nil
}

func (sender *proxySender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	handler := sender.handlers[eventHandler]
	if handler == nil {
		return errors.New("proxy events port not provided, cannot send events data")
	}

	if !handler.Connected() {
		if err := handler.Connect(); err != nil {
			return err
		}
	}

	line, err := EventLine(name, startMillis, endMillis, source, tags, setters...)
	if err != nil {
		return err
	}
	err = handler.SendData(line)
	return err
}

func (sender *proxySender) Close() {
	for _, h := range sender.handlers {
		if h != nil {
			h.Close()
		}
	}
}

func (sender *proxySender) Flush() error {
	errStr := ""
	for _, h := range sender.handlers {
		if h != nil {
			err := h.Flush()
			if err != nil {
				errStr = errStr + err.Error() + "\n"
			}
		}
	}
	if errStr != "" {
		return errors.New(strings.Trim(errStr, "\n"))
	}
	return nil
}

func (sender *proxySender) GetFailureCount() int64 {
	var failures int64
	for _, h := range sender.handlers {
		if h != nil {
			failures += h.GetFailureCount()
		}
	}
	return failures
}
