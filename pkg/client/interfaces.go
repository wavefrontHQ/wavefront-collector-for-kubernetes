package client

type WavefrontMetricSender interface {
	SendMetric(name, value, ts, source, tagStr string)
	Connect() error
	Close()
}
