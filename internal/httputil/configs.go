package httputil

import (
	"net/url"
)

type URL struct {
	*url.URL
}

// The configuration for a HTTP client.
type ClientConfig struct {
	// The bearer token for the client. Either this or the file below should be provided.
	BearerToken string `yaml:"bearer_token,omitempty"`
	// The bearer token file for the client. Either this or the token above should be provided.
	BearerTokenFile string `yaml:"bearer_token_file,omitempty"`
	// HTTP proxy server to use to connect.
	ProxyURL URL `yaml:"proxy_url,omitempty"`
	// TLSConfig to use to connect.
	TLSConfig TLSConfig `yaml:"tls_config,omitempty"`
}

// The TLS configuration for a HTTP client.
type TLSConfig struct {
	// The CA cert.
	CAFile string `yaml:"ca_file,omitempty"`
	// The client cert file.
	CertFile string `yaml:"cert_file,omitempty"`
	// The client key file.
	KeyFile string `yaml:"key_file,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `yaml:"server_name,omitempty"`
	// Disables certificate validation.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}
