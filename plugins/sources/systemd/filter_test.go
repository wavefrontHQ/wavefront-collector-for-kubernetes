// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package systemd

import (
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
)

func TestFromQuery(t *testing.T) {
	cfg := configuration.SystemdSourceConfig{}

	if fromConfig(cfg.UnitAllowList, cfg.UnitDenyList) != nil {
		t.Errorf("error creating filter")
	}

	// test allow lists
	cfg.UnitAllowList = []string{"docker*", "kubelet*"}
	f := fromConfig(cfg.UnitAllowList, cfg.UnitDenyList)
	if f == nil {
		t.Errorf("error creating filter")
	}
	if f.unitAllowList == nil {
		t.Errorf("error creating filter")
	}
	if !f.match("docker.service") {
		t.Errorf("error matching allowed docker.service")
	}
	if !f.match("kubelet.service") {
		t.Errorf("error matching allowed kubelet.service")
	}
	if f.match("random.service") {
		t.Errorf("error matching random.service")
	}

	// test deny lists
	cfg.UnitAllowList = nil
	cfg.UnitDenyList = []string{"*mount*", "etc*"}
	f = fromConfig(cfg.UnitAllowList, cfg.UnitDenyList)
	if f.match("home.mount") {
		t.Errorf("error matching denied home.mount")
	}
	if !f.match("kubelet.service") {
		t.Errorf("error matching kubelet.service")
	}
}
