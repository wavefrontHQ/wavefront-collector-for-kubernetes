package httputil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Returns a new HTTP client based on the given configuration.
func NewClient(cfg ClientConfig) (*http.Client, error) {
	rt, err := NewRoundTripper(cfg)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: rt}, nil
}

// Returns a new HTTP RoundTripper based on the given configuration.
func NewRoundTripper(cfg ClientConfig) (http.RoundTripper, error) {
	tlsConfig, err := NewTLSConfig(&cfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	var rt http.RoundTripper = &http.Transport{
		Proxy:               http.ProxyURL(cfg.ProxyURL.URL),
		MaxIdleConns:        20000,
		MaxIdleConnsPerHost: 1000, // see https://github.com/golang/go/issues/13801
		DisableKeepAlives:   false,
		TLSClientConfig:     tlsConfig,
		DisableCompression:  true,
		// dictates keepalive for connections.
		// 5 minutes is above the typical scrape interval for targets.
		IdleConnTimeout: 5 * time.Minute,
	}

	// create a round tripper that will set the Authz header if bearer token is provided
	if len(cfg.BearerToken) > 0 {
		rt = NewBearerTokenRoundTripper(cfg.BearerToken, rt)
	} else if len(cfg.BearerTokenFile) > 0 {
		rt = NewBearerTokenFileRoundTripper(cfg.BearerTokenFile, rt)
	}
	return rt, nil
}

type bearerTokenRoundTripper struct {
	token string
	rt    http.RoundTripper
}

// Returns a new HTTP RoundTripper that adds the given bearer token to a request header
func NewBearerTokenRoundTripper(token string, rt http.RoundTripper) http.RoundTripper {
	return &bearerTokenRoundTripper{token, rt}
}

func (rt *bearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) == 0 {
		req = cloneRequest(req)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(rt.token)))
	}
	return rt.rt.RoundTrip(req)
}

type bearerTokenFileRoundTripper struct {
	bearerFile string
	rt         http.RoundTripper
}

// Returns a new HTTP RoundTripper that adds the bearer token from the given file to a request header
func NewBearerTokenFileRoundTripper(bearerFile string, rt http.RoundTripper) http.RoundTripper {
	return &bearerTokenFileRoundTripper{bearerFile, rt}
}

func (rt *bearerTokenFileRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) == 0 {
		b, err := ioutil.ReadFile(rt.bearerFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read bearer token file %s: %s", rt.bearerFile, err)
		}
		bearerToken := strings.TrimSpace(string(b))

		req = cloneRequest(req)
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	return rt.rt.RoundTrip(req)
}

// cloneRequest returns a clone of the provided *http.Request.
func cloneRequest(r *http.Request) *http.Request {
	// Shallow copy of the struct.
	r2 := new(http.Request)
	*r2 = *r
	// Deep copy of the Header.
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
	}
	return r2
}

// Returns a new tls.Config from the given configuration.
func NewTLSConfig(cfg *TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}

	// If a CA cert is provided then let's read it in so we can validate the certificate.
	if len(cfg.CAFile) > 0 {
		caCertPool := x509.NewCertPool()
		// Load CA cert.
		caCert, err := ioutil.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("unable to use specified CA cert %s: %s", cfg.CAFile, err)
		}
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	if len(cfg.ServerName) > 0 {
		tlsConfig.ServerName = cfg.ServerName
	}
	// If a client cert & key is provided then configure TLS config accordingly.
	if len(cfg.CertFile) > 0 && len(cfg.KeyFile) == 0 {
		return nil, fmt.Errorf("client cert file %q specified without client key file", cfg.CertFile)
	} else if len(cfg.KeyFile) > 0 && len(cfg.CertFile) == 0 {
		return nil, fmt.Errorf("client key file %q specified without client cert file", cfg.KeyFile)
	} else if len(cfg.CertFile) > 0 && len(cfg.KeyFile) > 0 {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("unable to use specified client cert (%s) & key (%s): %s", cfg.CertFile, cfg.KeyFile, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	tlsConfig.BuildNameToCertificate()

	return tlsConfig, nil
}
