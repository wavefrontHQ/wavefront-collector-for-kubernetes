package senders

import (
	"fmt"
	"net/url"
	"strings"
)

// Option Wavefront client configuration options
type Option func(*configuration)

// Configuration for the direct ingestion sender
type configuration struct {
	Server string // Wavefront URL of the form https://<INSTANCE>.wavefront.com
	Token  string // Wavefront API token with direct data ingestion permission

	// Optional configuration properties. Default values should suffice for most use cases.
	// override the defaults only if you wish to set higher values.

	// max batch of data sent per flush interval. defaults to 10,000. recommended not to exceed 40,000.
	BatchSize int

	// size of internal buffers beyond which received data is dropped.
	// helps with handling brief increases in data and buffering on errors.
	// separate buffers are maintained per data type (metrics, spans and distributions)
	// buffers are not pre-allocated to max size and vary based on actual usage.
	// defaults to 500,000. higher values could use more memory.
	MaxBufferSize int

	// interval (in seconds) at which to flush data to Wavefront. defaults to 1 Second.
	// together with batch size controls the max theoretical throughput of the sender.
	FlushIntervalSeconds int
}

// NewSender creates Wavefront client
func NewSender(wfURL string, setters ...Option) (Sender, error) {
	cfg := &configuration{}

	u, err := url.Parse(wfURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(strings.ToLower(u.Scheme), "http") {
		return nil, fmt.Errorf("invalid schema '%s', only 'http' is supported", u.Scheme)
	}

	if len(u.User.String()) > 0 {
		cfg.Token = u.User.String()
		u.User = nil
	}

	cfg.Server = u.String()

	for _, set := range setters {
		set(cfg)
	}
	return newWavefrontClient(cfg)
}

// BatchSize set max batch of data sent per flush interval. defaults to 10,000. recommended not to exceed 40,000.
func BatchSize(n int) Option {
	return func(cfg *configuration) {
		cfg.BatchSize = n
	}
}

// MaxBufferSize set the size of internal buffers beyond which received data is dropped.
func MaxBufferSize(n int) Option {
	return func(cfg *configuration) {
		cfg.MaxBufferSize = n
	}
}

// FlushIntervalSeconds set the interval (in seconds) at which to flush data to Wavefront. defaults to 1 Second.
func FlushIntervalSeconds(n int) Option {
	return func(cfg *configuration) {
		cfg.FlushIntervalSeconds = n
	}
}
