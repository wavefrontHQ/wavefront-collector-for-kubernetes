package main

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"strings"
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/options"
)

func TestAddQueryKey(t *testing.T) {
	u, _ := url.Parse("?")
	kuri := flags.Uri{Key: "kubernetes.summary_api", Val: *u}

	uri := addQueryKey(kuri, "prefix", "staging.")
	assert.True(t, strings.Contains(uri.String(), "prefix=staging."))
}

func TestRemoveQueryKey(t *testing.T) {
	values := url.Values{}
	values.Add("proxyAddress", "localhost:2878")
	values.Add("prefix", "staging.")
	u, _ := url.Parse("?")
	u.RawQuery = values.Encode()
	suri := flags.Uri{Key: "wavefront", Val: *u}

	assert.True(t, strings.Contains(suri.String(), "prefix=staging."))
	uri := removeQueryKey(suri, "prefix")
	assert.True(t, !strings.Contains(uri.String(), "prefix=staging."))
}

func TestHandleSinkPrefix(t *testing.T) {
	opts := options.NewCollectorRunOptions()

	values := url.Values{}
	values.Add("proxyAddress", "localhost:2878")
	values.Add("prefix", "staging.")

	u, err := url.Parse("?")
	if err != nil {
		t.Error(err)
	}
	u.RawQuery = values.Encode()
	uri := flags.Uri{Key: "wavefront", Val: *u}
	opts.Sinks = append(opts.Sinks, uri)

	assert.True(t, strings.Contains(opts.Sinks[0].String(), "prefix=staging."))

	u, _ = url.Parse("?")
	kuri := flags.Uri{Key: "kubernetes.summary_api", Val: *u}
	opts.Sources = append(opts.Sources, kuri)

	assert.True(t, !strings.Contains(opts.Sources[0].String(), "prefix=staging."))
	cleanupSinkPrefix(opts)
	assert.True(t, strings.Contains(opts.Sources[0].String(), "prefix=staging."))
}
