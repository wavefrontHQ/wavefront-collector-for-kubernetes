package senders

type SpanTag struct {
	Key   string
	Value string
}

type SpanLog struct {
	Timestamp int64             `json:"timestamp"`
	Fields    map[string]string `json:"fields"`
}

type SpanLogs struct {
	TraceId string    `json:"traceId"`
	SpanId  string    `json:"spanId"`
	Logs    []SpanLog `json:"logs"`
}
