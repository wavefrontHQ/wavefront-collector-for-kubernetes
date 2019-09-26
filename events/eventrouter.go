package events

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var Log = log.WithField("system", "events")
var leadershipName = "wf_collector_events"

type EventRouter struct {
	kubeClient        kubernetes.Interface
	eLister           corelisters.EventLister
	eListerSynched    cache.InformerSynced
	skins             []metrics.DataSink
	sharedInformers   informers.SharedInformerFactory
	stop              chan struct{}
	daemon            bool
	leadershipManager *leadershipManager
}

type EventSinkInterface interface {
	UpdateEvents(function string, eNew *v1.Event, eOld *v1.Event)
}

func CreateEventRouter(clientset kubernetes.Interface, skins []metrics.DataSink, clusterName string, daemon bool) *EventRouter {
	sharedInformers := informers.NewSharedInformerFactory(clientset, time.Minute)
	eventsInformer := sharedInformers.Core().V1().Events()

	er := &EventRouter{
		kubeClient:      clientset,
		skins:           skins,
		daemon:          daemon,
		sharedInformers: sharedInformers,
	}
	eventsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: er.addEvent,
		// UpdateFunc: er.updateEvent,
		// DeleteFunc: er.deleteEvent,
	})
	er.eLister = eventsInformer.Lister()
	er.eListerSynched = eventsInformer.Informer().HasSynced

	if er.daemon {
		er.leadershipManager = newLeadershipManager(er, leadershipName, clientset)
	}
	return er
}

func (er *EventRouter) Start() {
	if er.daemon {
		er.leadershipManager.Start()
	} else {
		go func() { er.resume() }()
	}
}

func (er *EventRouter) resume() {
	er.stop = make(chan struct{})
	defer utilruntime.HandleCrash()

	Log.Infof("Starting EventRouter")

	go func() { er.sharedInformers.Start(er.stop) }()

	// here is where we kick the caches into gear
	if !cache.WaitForCacheSync(er.stop, er.eListerSynched) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	<-er.stop

	Log.Infof("Shutting down EventRouter")
}

func (er *EventRouter) pause() {
	close(er.stop)
}

func (er *EventRouter) Stop() {
	if er.daemon {
		er.leadershipManager.Stop()
	}
	er.pause()
}

// addEvent is called when an event is created, or during the initial list
func (er *EventRouter) addEvent(obj interface{}) {
	e := obj.(*v1.Event)
	for _, skin := range er.skins {
		skin.ExportEvents("added", e, nil)
	}
}

// updateEvent is called any time there is an update to an existing event
func (er *EventRouter) updateEvent(objOld interface{}, objNew interface{}) {
	eOld := objOld.(*v1.Event)
	eNew := objNew.(*v1.Event)
	for _, skin := range er.skins {
		skin.ExportEvents("update", eNew, eOld)
	}
}

// deleteEvent should only occur when the system garbage collects events via TTL expiration
func (er *EventRouter) deleteEvent(obj interface{}) {
	e := obj.(*v1.Event)
	for _, skin := range er.skins {
		skin.ExportEvents("delete", e, nil)
	}
}

type system interface {
	resume()
	pause()
}

type leadershipManager struct {
	system     system
	name       string
	stop       chan struct{}
	kubeClient kubernetes.Interface
}

func newLeadershipManager(system system, name string, kubeClient kubernetes.Interface) *leadershipManager {
	return &leadershipManager{
		stop:       make(chan struct{}),
		system:     system,
		name:       name,
		kubeClient: kubeClient,
	}
}

func (lm *leadershipManager) Start() {
	ch, err := leadership.Subscribe(lm.kubeClient.CoreV1(), lm.name)
	if err != nil {
		Log.Errorf("discovery: leader election error: %q", err)
	} else {
		go func() { lm.run(ch) }()
	}
}

func (lm *leadershipManager) Stop() {
	close(lm.stop)
	leadership.Unsubscribe(lm.name)
}

func (lm *leadershipManager) run(ch <-chan bool) {
	for {
		select {
		case isLeader := <-ch:
			if isLeader {
				Log.Infof("promoted to leader for '%v' node:'%s'", lm.name, leadership.Leader())
				go func() { lm.system.resume() }()
			} else {
				Log.Infof("demoted from leader events for '%v' new leader:'%s'", lm.name, leadership.Leader())
				lm.system.pause()
			}
		case <-lm.stop:
			Log.Infof("stopping leadershipManager for '%v'", lm.name)
			return
		}
	}
}
