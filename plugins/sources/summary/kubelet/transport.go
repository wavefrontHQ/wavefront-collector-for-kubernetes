// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kubelet

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

type KubeletClientConfig struct {
	// Default port - used if no information about Kubelet port can be found in Node.NodeStatus.DaemonEndpoints.
	Port         uint
	ReadOnlyPort uint
	EnableHttps  bool

	// PreferredAddressTypes - used to select an address from Node.NodeStatus.Addresses
	PreferredAddressTypes []string

	// TLSClientConfig contains settings to enable transport layer security
	restclient.TLSClientConfig

	// Server requires Bearer authentication
	BearerToken string

	// HTTPTimeout is used by the client to timeout http requests to Kubelet.
	HTTPTimeout time.Duration

	// Dial is a custom dialer used for the client
	Dial utilnet.DialFunc
}

func (c *KubeletClientConfig) HTTPSEnabled() bool {
	if c == nil {
		return true
	}
	return c.EnableHttps
}

func (c *KubeletClientConfig) Scheme() string {
	if c.HTTPSEnabled() {
		return "https"
	}
	return "http"
}

func (c *KubeletClientConfig) BaseURL(ip net.IP, path string) *url.URL {
	return &url.URL{
		Scheme: c.Scheme(),
		Host:   fmt.Sprintf("%s:%d", ip, c.Port),
		Path:   path,
	}
}

func MakeTransport(config *KubeletClientConfig) (http.RoundTripper, error) {
	tlsConfig, err := transport.TLSConfigFor(config.transportConfig())
	if err != nil {
		return nil, err
	}

	rt := http.DefaultTransport
	if config.Dial != nil || tlsConfig != nil {
		rt = utilnet.SetOldTransportDefaults(&http.Transport{
			DialContext:     config.Dial,
			TLSClientConfig: tlsConfig,
		})
	}

	return transport.HTTPWrappersForConfig(config.transportConfig(), rt)
}

// transportConfig converts a client config to an appropriate transport config.
func (c *KubeletClientConfig) transportConfig() *transport.Config {
	cfg := &transport.Config{
		TLS: transport.TLSConfig{
			CAFile:   c.CAFile,
			CAData:   c.CAData,
			CertFile: c.CertFile,
			CertData: c.CertData,
			KeyFile:  c.KeyFile,
			KeyData:  c.KeyData,
		},
	}
	if c.EnableHttps {
		cfg.BearerToken = c.BearerToken
	}
	if c.EnableHttps && !cfg.HasCA() {
		cfg.TLS.Insecure = true
	}
	return cfg
}
