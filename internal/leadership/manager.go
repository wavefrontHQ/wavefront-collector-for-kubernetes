// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package leadership

import (
	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
)

type Resumer interface {
	Resume()
	Pause()
}

// Manager manages the pausing and resumption of a given system based on changes in leadership
type Manager struct {
	system     Resumer
	name       string
	stop       chan struct{}
	kubeClient kubernetes.Interface
}

// NewManager creates a new leadership manager for a given system
func NewManager(system Resumer, name string, kubeClient kubernetes.Interface) *Manager {
	return &Manager{
		stop:       make(chan struct{}),
		system:     system,
		name:       name,
		kubeClient: kubeClient,
	}
}

func (lm *Manager) Start() {
	ch, err := Subscribe(lm.kubeClient.CoreV1(), lm.name)
	if err != nil {
		log.Errorf("%s: leader election error: %q", lm.name, err)
	} else {
		go func() {
			lm.run(ch)
		}()
	}
}

func (lm *Manager) Stop() {
	close(lm.stop)
	Unsubscribe(lm.name)
}

func (lm *Manager) run(ch <-chan bool) {
	for {
		select {
		case isLeader := <-ch:
			if isLeader {
				log.Infof("resuming %s: node %s elected leader", lm.name, Leader())
				go func() { lm.system.Resume() }()
			} else {
				log.Infof("pausing %s: demoted from leadership. new leader: %s", lm.name, Leader())
				lm.system.Pause()
			}
		case <-lm.stop:
			log.Infof("%s: stopping leadership manager", lm.name)
			return
		}
	}
}
