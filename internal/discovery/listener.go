package discovery

import (
	"github.com/golang/glog"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
)

type listener struct {
	handler util.ConfigHandler
}

func NewFileListener(handler util.ConfigHandler) util.FileListener {
	return &listener{handler: handler}
}

func (l *listener) Changed(file string) {
	cfg, err := FromFile(file)
	if err != nil {
		glog.Errorf("error loading discovery config: %v", err)
	} else {
		l.handler.Handle(cfg)
	}
}
