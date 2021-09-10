package application

import (
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

// HeartbeatService sends a heartbeat metric every 5 minutes
type HeartbeatService interface {
	Close()
	AddCustomTags(tags map[string]string)
}

type heartbeater struct {
	sender      senders.Sender
	application Tags
	source      string
	components  []string
	mux         sync.Mutex
	customTags  []map[string]string

	ticker *time.Ticker
	stop   chan struct{}
}

// StartHeartbeatService will create and start a new HeartbeatService
func StartHeartbeatService(sender senders.Sender, application Tags, source string, components ...string) HeartbeatService {
	hb := &heartbeater{
		sender:      sender,
		application: application,
		source:      source,
		components:  components,
		ticker:      time.NewTicker(5 * time.Minute),
		stop:        make(chan struct{}),
	}
	go func() {
		for {
			select {
			case <-hb.ticker.C:
				hb.beat()
			case <-hb.stop:
				return
			}
		}
	}()

	hb.beat()
	return hb
}

func (hb *heartbeater) Close() {
	hb.ticker.Stop()
	hb.stop <- struct{}{} // block until goroutine exits
}

func (hb *heartbeater) beat() {
	tags := hb.application.Map()
	tags["component"] = "wavefront-generated"
	hb.send(tags)

	for _, component := range hb.components {
		tags["component"] = component
		hb.send(tags)
	}

	//send customTags heartbeat
	hb.mux.Lock()
	for len(hb.customTags) > 0 {
		tags := hb.customTags[0]
		hb.customTags = hb.customTags[1:]
		hb.send(tags)
	}
	hb.mux.Unlock()
}

func (hb *heartbeater) send(tags map[string]string) {
	err := hb.sender.SendMetric("~component.heartbeat", 1, 0, hb.source, tags)
	if err != nil {
		log.Printf("heartbeater SendMetric error: %v\n", err)
	}
}

func (hb *heartbeater) AddCustomTags(tags map[string]string) {
	hb.mux.Lock()
	defer hb.mux.Unlock()
	for _, existCustomTag := range hb.customTags {
		if reflect.DeepEqual(existCustomTag, tags) {
			return
		}
	}

	newTags := make(map[string]string)

	for k, v := range tags {
		newTags[k] = v
	}

	hb.customTags = append(hb.customTags, newTags)
}
