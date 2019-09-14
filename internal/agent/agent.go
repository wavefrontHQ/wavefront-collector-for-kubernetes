// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/informers"

	"github.com/wavefronthq/wavefront-kubernetes-collector/events"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/manager"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources"
)

type Agent struct {
	pm   manager.FlushManager
	dm   *discovery.Manager
	er   *events.EventRouter
	sif  informers.SharedInformerFactory
	stop chan struct{}
}

func NewAgent(pm manager.FlushManager, dm *discovery.Manager, er *events.EventRouter, sif informers.SharedInformerFactory) *Agent {
	return &Agent{
		pm:   pm,
		dm:   dm,
		er:   er,
		sif:  sif,
		stop: make(chan struct{}),
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
	log.Infof("Starting Events collector")
	go func() { a.er.Run(a.stop) }()
	go func() { a.sif.Start(a.stop) }()
	log.Infof("Done Starting Events collector")
}

func (a *Agent) Stop() {
	log.Infof("Stopping agent")
	close(a.stop)
	a.pm.Stop()
	if a.dm != nil {
		a.dm.Stop()
	}
	sources.Manager().StopProviders()
	log.Infof("Agent stopped")
}
