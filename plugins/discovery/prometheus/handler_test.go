// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
)

func TestDiscover(t *testing.T) {
	th := NewTargetHandler(true, &util.DummyProviderHandler{}, "prometheus.io")

	pod := discovery.FakePod("pod1", "ns", "123")
	th.Handle(discovery.Resource{
		IP:   "123",
		Kind: discovery.PodType.String(),
		Meta: pod.ObjectMeta,
	}, discovery.PluginConfig{})

	// should not be discovered without annotations
	if th.Count() != 0 {
		t.Error("unexpected pod1 registration")
	}

	// add annotations and discover again
	pod.Annotations = map[string]string{
		"prometheus.io/scrape": "true",
	}
	th.Handle(discovery.Resource{
		IP:   "123",
		Kind: discovery.PodType.String(),
		Meta: pod.ObjectMeta,
	}, discovery.PluginConfig{})

	// should be discovered
	if th.Count() != 1 {
		t.Error("expected pod1 registration")
	}
}
