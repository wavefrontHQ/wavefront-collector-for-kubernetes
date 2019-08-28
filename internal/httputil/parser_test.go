// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package httputil

import "testing"

var sampleConf = `
bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token'
tls_config:
  ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'
  insecure_skip_verify: true
`

func TestFromYAML(t *testing.T) {
	cfg, err := FromYAML([]byte(sampleConf))
	if err != nil {
		t.Errorf("error loading yaml: %q", err)
		return
	}
	if cfg.BearerTokenFile != "/var/run/secrets/kubernetes.io/serviceaccount/token" {
		t.Errorf("invalid bearer token file: %s", cfg.BearerTokenFile)
	}
	if cfg.TLSConfig.CAFile != "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt" {
		t.Errorf("invalid tls CA file: %s", cfg.TLSConfig.CAFile)
	}
	if !cfg.TLSConfig.InsecureSkipVerify {
		t.Errorf("invalid tls insecure")
	}
}
