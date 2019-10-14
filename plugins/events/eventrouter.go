package events

import (
	"fmt"
	"time"

	"github.com/gobwas/glob"
	gometrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sinks"
	"github.com/wavefronthq/wavefront-sdk-go/event"
	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var Log = log.WithField("system", "events")
var leadershipName = "wf_collector_events"
var filteredEvents = gometrics.GetOrRegisterCounter("events.filtered", gometrics.DefaultRegistry)
var receviedEvents = gometrics.GetOrRegisterCounter("events.recevied", gometrics.DefaultRegistry)
var sentEvents = gometrics.GetOrRegisterCounter("events.sent", gometrics.DefaultRegistry)

type EventRouter struct {
	kubeClient        kubernetes.Interface
	eLister           corelisters.EventLister
	eListerSynched    cache.InformerSynced
	skins             []metrics.DataSink
	sharedInformers   informers.SharedInformerFactory
	stop              chan struct{}
	daemon            bool
	leadershipManager *leadershipManager
	whitelist         map[string]glob.Glob
	blacklist         map[string]glob.Glob
}

func CreateEventRouter(clientset kubernetes.Interface, cfg *configuration.EventsConfig, clusterName string, daemon bool) *EventRouter {
	sharedInformers := informers.NewSharedInformerFactory(clientset, time.Minute)
	eventsInformer := sharedInformers.Core().V1().Events()

	sinksFactory := sinks.NewSinkFactory()
	skins := sinksFactory.BuildAll(cfg.Sinks, false)
	if len(skins) == 0 {
		log.Fatalf("Failed to create Event manager manager: no Sink defined")
	}

	er := &EventRouter{
		kubeClient:      clientset,
		skins:           skins,
		daemon:          daemon,
		sharedInformers: sharedInformers,
		whitelist:       filter.MultiCompile(cfg.Filters.TagWhitelist),
		blacklist:       filter.MultiCompile(cfg.Filters.TagBlacklist),
	}

	eventsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: er.addEvent,
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

	ns := e.InvolvedObject.Namespace
	if len(ns) == 0 {
		ns = "default"
	}

	tags := map[string]string{
		"namespace": ns,
		"Kind":      e.InvolvedObject.Kind,
		"Reason":    e.Reason,
		"component": e.Source.Component,
	}

	receviedEvents.Inc(1)
	if len(er.whitelist) > 0 && !filter.MatchesTags(er.whitelist, tags) {
		Log.Infof("event '%s' filtered becuase a white list", e.Message)
		filteredEvents.Inc(1)
		return
	}
	if len(er.blacklist) > 0 && filter.MatchesTags(er.blacklist, tags) {
		Log.Infof("event '%s' filtered becuase a black list", e.Message)
		filteredEvents.Inc(1)
		return
	}
	sentEvents.Inc(1)

	eType := e.Type
	if len(eType) == 0 {
		eType = "Normal"
	}

	for _, skin := range er.skins {
		skin.ExportEvents(
			e.Message,
			e.LastTimestamp.Time,
			e.Source.Host,
			tags,
			event.Type(eType),
		)
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
