// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package leadership

import (
	"fmt"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

var (
	// internal metrics
	electionError metrics.Counter
	leadingGauge  metrics.Gauge

	// leadership state
	subscribers map[string]chan<- bool
	lock        sync.RWMutex
	started     bool
	isLeader    bool
	leaderId    string
)

func init() {
	electionError = metrics.GetOrRegisterCounter("leaderelection.error", metrics.DefaultRegistry)
	leadingGauge = metrics.GetOrRegisterGauge("leaderelection.leading", metrics.DefaultRegistry)
}

// Subscribe starts the leader election process if not already started
// and returns a channel subscriber can listen on for election results
func Subscribe(client v1.CoreV1Interface, name string) (<-chan bool, error) {
	lock.Lock()
	defer lock.Unlock()

	if err := startLeaderElection(client); err != nil {
		return nil, err
	}
	ch := make(chan bool, 1)
	// inform if we are currently the leader
	if isLeader {
		ch <- true
	}
	if subscribers == nil {
		subscribers = make(map[string]chan<- bool)
	}
	// add to subscribers map to notify of election results
	subscribers[name] = ch
	return ch, nil
}

func Unsubscribe(name string) {
	lock.Lock()
	defer lock.Unlock()

	delete(subscribers, name)
	log.Infof("unsubscribed %s from leader-election: %d", name, len(subscribers))
}

// startLeaderElection starts the election process if not already started
// this will only be done once per collector instance
func startLeaderElection(client v1.CoreV1Interface) error {
	if !started {
		le, err := getLeaderElector(client)
		if err != nil {
			electionError.Inc(1)
			return err
		}
		go func() {
			for {
				le.Run()
			}
		}()
		started = true
	}
	return nil
}

// getLeaderElector returns a leader elector
func getLeaderElector(client v1.CoreV1Interface) (*leaderelection.LeaderElector, error) {
	nodeName := util.GetNodeName()
	if nodeName == "" {
		return nil, fmt.Errorf("%s envvar is not defined", util.NodeNameEnvVar)
	}
	ns := util.GetNamespaceName()
	if ns == "" {
		return nil, fmt.Errorf("%s envvar is not defined", util.NamespaceNameEnvVar)
	}

	resourceLock, err := getResourceLock(ns, "wf-collector-leader", client, nodeName)
	if err != nil {
		return nil, err
	}

	le, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          resourceLock,
		LeaseDuration: 60 * time.Second,
		RenewDeadline: 45 * time.Second,
		RetryPeriod:   30 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(stop <-chan struct{}) {},
			OnStoppedLeading: func() {},
			OnNewLeader: func(identity string) {
				lock.Lock()
				defer lock.Unlock()

				if identity == nodeName {
					leadingGauge.Update(1)
				} else {
					leadingGauge.Update(0)
				}

				log.Infof("node: %s elected leader", identity)
				leaderId = identity
				if identity == nodeName && !isLeader {
					for i := range subscribers {
						subscribers[i] <- true
					}
					isLeader = true
				} else if identity != nodeName && isLeader {
					for i := range subscribers {
						subscribers[i] <- false
					}
				}
			},
		},
	})
	return le, err
}

// getResourceLock returns a config map based resource lock for leader election
func getResourceLock(ns string, name string, client v1.CoreV1Interface, resourceLockID string) (resourcelock.Interface, error) {
	return resourcelock.New(
		resourcelock.ConfigMapsResourceLock,
		ns,
		name,
		client,
		resourcelock.ResourceLockConfig{
			Identity:      resourceLockID,
			EventRecorder: &record.FakeRecorder{},
		},
	)
}

func Leader() string {
	lock.RLock()
	defer lock.RUnlock()
	return leaderId
}

func Leading() bool {
	lock.RLock()
	defer lock.RUnlock()
	return util.GetDaemonMode() == "" || isLeader
}
