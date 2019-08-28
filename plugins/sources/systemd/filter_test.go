// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package systemd

import (
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
)

func TestFromQuery(t *testing.T) {
	cfg := configuration.SystemdSourceConfig{}

	if fromConfig(cfg.UnitWhitelist, cfg.UnitBlacklist) != nil {
		t.Errorf("error creating filter")
	}

	// test whitelisting
	cfg.UnitWhitelist = []string{"docker*", "kubelet*"}
	f := fromConfig(cfg.UnitWhitelist, cfg.UnitBlacklist)
	if f == nil {
		t.Errorf("error creating filter")
	}
	if f.unitWhitelist == nil {
		t.Errorf("error creating filter")
	}
	if !f.match("docker.service") {
		t.Errorf("error matching whitelisted docker.service")
	}
	if !f.match("kubelet.service") {
		t.Errorf("error matching whitelisted kubelet.service")
	}
	if f.match("random.service") {
		t.Errorf("error matching random.service")
	}

	// test blacklisting
	cfg.UnitWhitelist = nil
	cfg.UnitBlacklist = []string{"*mount*", "etc*"}
	f = fromConfig(cfg.UnitWhitelist, cfg.UnitBlacklist)
	if f.match("home.mount") {
		t.Errorf("error matching blacklisted home.mount")
	}
	if !f.match("kubelet.service") {
		t.Errorf("error matching kubelet.service")
	}
}
