// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/events"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/manager"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources"
)

type Agent struct {
	pm manager.FlushManager
	dm *discovery.Manager
	er *events.EventRouter
}

func NewAgent(pm manager.FlushManager, dm *discovery.Manager, er *events.EventRouter) *Agent {
	return &Agent{
		pm: pm,
		dm: dm,
		er: er,
	}
}

func (a *Agent) Start() {
	log.Infof("Starting agent")
	a.pm.Start()
	if a.dm != nil {
		a.dm.Start()
	}

	if a.er != nil {
		log.Infof("Starting Events collector")
		a.er.Start()
		log.Infof("Done Starting Events collector")
	}
}

func (a *Agent) Stop() {
	log.Infof("Stopping agent")
	a.pm.Stop()
	if a.dm != nil {
		a.dm.Stop()
	}

	if a.er != nil {
		log.Infof("Stopping Events collector")
		a.er.Stop()
		log.Infof("Done Stopping Events collector")
	}

	sources.Manager().StopProviders()
	log.Infof("Agent stopped")
}
