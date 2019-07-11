package agent

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/manager"
)

type Agent struct {
	mgr manager.Manager
	dm  *discovery.Manager
}

func NewAgent(mgr manager.Manager, dm *discovery.Manager) *Agent {
	return &Agent{
		mgr: mgr,
		dm:  dm,
	}
}

func (a *Agent) Handle(cfg interface{}) {
	if a.dm != nil {
		a.dm.Handle(cfg)
	}
}

func (a *Agent) Start() {
	a.mgr.Start()
	if a.dm != nil {
		a.dm.Start()
	}
}

func (a *Agent) Stop() {
	// not supported for now
}
