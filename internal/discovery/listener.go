// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
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
		log.Errorf("error loading discovery config: %v", err)
	} else {
		l.handler.Handle(cfg)
	}
}
