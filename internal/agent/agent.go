// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/manager"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources"
)

type Agent struct {
	pm manager.FlushManager
	dm *discovery.Manager
}

func NewAgent(pm manager.FlushManager, dm *discovery.Manager) *Agent {
	return &Agent{
		pm: pm,
		dm: dm,
	}
}

func (a *Agent) Handle(cfg interface{}) {
	if a.dm != nil {
		a.dm.Handle(cfg)
	}
}

func (a *Agent) Start() {
	log.Infof("Starting agent")
	a.pm.Start()
	if a.dm != nil {
		a.dm.Start()
	}
}

func (a *Agent) Stop() {
	log.Infof("Stopping agent")
	a.pm.Stop()
	if a.dm != nil {
		a.dm.Stop()
	}
	sources.Manager().StopProviders()
	log.Infof("Agent stopped")
}
