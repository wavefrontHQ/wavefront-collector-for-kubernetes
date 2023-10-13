package senders

import (
	"crypto/tls"
	"github.com/wavefronthq/wavefront-sdk-go/internal/auth"
	"time"
)

// Option Wavefront client configuration options
type Option func(*configuration)

// APIToken configures the sender to use a Wavefront API Token for authentication
func APIToken(apiToken string) Option {
	return func(c *configuration) {
		c.Authentication = auth.APIToken{
			Token: apiToken,
		}
	}
}

// A CSPOption sets optional configuration for CSP Authentication
type CSPOption func(any)

// CSPBaseURL sets an alternative base URL for the CSP server
func CSPBaseURL(baseURL string) CSPOption {
	return func(authentication any) {
		switch authentication.(type) {
		case *auth.CSPClientCredentials:
			credentials := authentication.(*auth.CSPClientCredentials)
			credentials.BaseURL = baseURL
		case *auth.CSPAPIToken:
			token := authentication.(*auth.CSPAPIToken)
			token.BaseURL = baseURL
		}
	}
}

// CSPOrgID sets an explicit orgID for Client Credentials authentication
func CSPOrgID(orgID string) CSPOption {
	return func(authentication any) {
		switch authentication.(type) {
		case auth.CSPClientCredentials:
			credentials := authentication.(auth.CSPClientCredentials)
			credentials.OrgID = &orgID
		}
	}
}

// CSPAPIToken configures the sender to use a CSP API Token for authentication
func CSPAPIToken(cspAPIToken string, options ...CSPOption) Option {
	return func(c *configuration) {
		cspTokenAuth := auth.CSPAPIToken{
			Token:   cspAPIToken,
			BaseURL: defaultCSPBaseUrl,
		}
		for _, option := range options {
			option(&cspTokenAuth)
		}
		c.Authentication = cspTokenAuth
	}
}

// CSPClientCredentials configures the sender to use a CSP Client Credentials for authentication
func CSPClientCredentials(clientId string, clientSecret string, options ...CSPOption) Option {
	return func(c *configuration) {
		clientCredentials := &auth.CSPClientCredentials{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			BaseURL:      defaultCSPBaseUrl,
		}
		for _, option := range options {
			option(clientCredentials)
		}
		c.Authentication = *clientCredentials
	}
}

// BatchSize set max batch of data sent per flush interval. Defaults to 10,000. recommended not to exceed 40,000.
func BatchSize(n int) Option {
	return func(cfg *configuration) {
		cfg.BatchSize = n
	}
}

// MaxBufferSize set the size of internal buffers beyond which received data is dropped. Defaults to 50,000.
func MaxBufferSize(n int) Option {
	return func(cfg *configuration) {
		cfg.MaxBufferSize = n
	}
}

// FlushIntervalSeconds set the interval (in seconds) at which to flush data to Wavefront. Defaults to 1 Second.
func FlushIntervalSeconds(n int) Option {
	return func(cfg *configuration) {
		cfg.FlushInterval = time.Second * time.Duration(n)
	}
}

// FlushInterval set the interval at which to flush data to Wavefront. Defaults to 1 Second.
func FlushInterval(interval time.Duration) Option {
	return func(cfg *configuration) {
		cfg.FlushInterval = interval
	}
}

// MetricsPort sets the port on which to report metrics. Default is 2878.
func MetricsPort(port int) Option {
	return func(cfg *configuration) {
		cfg.MetricsPort = port
	}
}

// TracesPort sets the port on which to report traces. Default is 30001.
func TracesPort(port int) Option {
	return func(cfg *configuration) {
		cfg.TracesPort = port
	}
}

// Timeout sets the HTTP timeout (in seconds). Defaults to 10 seconds.
func Timeout(timeout time.Duration) Option {
	return func(cfg *configuration) {
		cfg.Timeout = timeout
	}
}

// TLSConfigOptions sets the tls.Config used by the HTTP Client to send data to Wavefront.
func TLSConfigOptions(tlsCfg *tls.Config) Option {
	tlsCfgCopy := tlsCfg.Clone()
	return func(cfg *configuration) {
		cfg.TLSConfig = tlsCfgCopy
	}
}

// SendInternalMetrics turns sending of internal SDK metrics on/off.
func SendInternalMetrics(enabled bool) Option {
	return func(cfg *configuration) {
		cfg.SendInternalMetrics = enabled
	}
}

// SDKMetricsTags adds the additional tags provided in tags to all internal
// metrics this library reports. Clients can use multiple SDKMetricsTags
// calls when creating a sender. In that case, the sender sends all the
// tags from each of the SDKMetricsTags calls in addition to the standard
// "pid" and "version" tags to all internal metrics. The "pid" tag is the
// process ID; the "version" tag is the version of this SDK.
func SDKMetricsTags(tags map[string]string) Option {
	// prevent caller from accidentally mutating this option.
	copiedTags := copyTags(tags)
	return func(cfg *configuration) {
		for key, value := range copiedTags {
			cfg.SDKMetricsTags[key] = value
		}
	}
}
