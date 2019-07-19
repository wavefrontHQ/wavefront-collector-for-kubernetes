package agent

import (
	"github.com/golang/glog"
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
	glog.Infof("Starting agent")
	a.pm.Start()
	if a.dm != nil {
		a.dm.Start()
	}
}

func (a *Agent) Stop() {
	glog.Infof("Stopping agent")
	a.pm.Stop()
	if a.dm != nil {
		a.dm.Stop()
	}
	sources.Manager().StopProviders()
	glog.Infof("Agent stopped")
}
