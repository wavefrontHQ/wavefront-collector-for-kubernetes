// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"reflect"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

type ProviderInfo struct {
	Handler metrics.ProviderHandler
	Factory metrics.ProviderFactory
	Encoder Encoder
}

type defaultHandler struct {
	info           ProviderInfo
	registry       TargetRegistry
	rh             func(r Resource) bool
	useAnnotations bool

	mtx     sync.RWMutex
	targets map[string]interface{}
}

type HandlerOption func(TargetHandler)

func UseAnnotations(use bool) HandlerOption {
	return func(handler TargetHandler) {
		if h, ok := handler.(*defaultHandler); ok {
			h.useAnnotations = use
		}
	}
}

func SetRegistrationHandler(f func(resource Resource) bool) HandlerOption {
	return func(handler TargetHandler) {
		if h, ok := handler.(*defaultHandler); ok {
			h.rh = f
		}
	}
}

// Gets a new target handler for handling discovered targets
func NewHandler(info ProviderInfo, registry TargetRegistry, setters ...HandlerOption) TargetHandler {
	handler := &defaultHandler{
		info:     info,
		registry: registry,
		targets:  make(map[string]interface{}),
	}
	for _, setter := range setters {
		setter(handler)
	}
	return handler
}

func (d *defaultHandler) Encoding(name string) interface{} {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return d.targets[name]
}

func (d *defaultHandler) add(name string, cfg interface{}) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.targets[name] = cfg
}

func (d *defaultHandler) Delete(name string) {
	d.unregister(name)
}

// deletes targets that do not exist in the input map
func (d *defaultHandler) DeleteMissing(input map[string]bool) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	for k := range d.targets {
		if _, exists := input[k]; !exists {
			// delete directly rather than call unregister to prevent recursive locking
			delete(d.targets, k)
			d.deleteProvider(k)
		}
	}
}

func (d *defaultHandler) Handle(resource Resource, rule interface{}) {
	kind := resource.Kind
	ip := resource.IP
	meta := resource.Meta

	log.WithFields(log.Fields{
		"kind":      kind,
		"name":      meta.Name,
		"namespace": meta.Namespace,
	}).Debug("Handling resource")

	name := ResourceName(kind, meta)
	currEncoding := d.registry.Encoding(name)
	newEncoding, ok := d.info.Encoder.Encode(ip, kind, meta, rule)

	// add target if newEncoding is non-empty and has changed
	if ok && !reflect.DeepEqual(currEncoding, newEncoding) {
		log.Debugf("newEncoding: %s", newEncoding)
		log.Debugf("currEncoding: %s", currEncoding)

		provider, err := d.info.Factory.Build(newEncoding)
		if err != nil {
			log.Error(err)
			return
		}
		d.register(name, newEncoding, provider)
	}

	// delete target if scrape annotation is false/absent and handler is annotation based
	if !ok && currEncoding != nil && d.useAnnotations && d.Encoding(name) != nil {
		if d.rh != nil && d.rh(resource) {
			log.Infof("deleting target %s as annotation has changed", name)
			d.unregister(name)
		}
	}
}

func (d *defaultHandler) register(name string, cfg interface{}, provider metrics.MetricsSourceProvider) {
	d.add(name, cfg)
	d.info.Handler.AddProvider(provider)
	d.registry.Register(name, d)
}

func (d *defaultHandler) unregister(name string) {
	d.mtx.Lock()
	delete(d.targets, name)
	d.mtx.Unlock()
	d.deleteProvider(name)
}

func (d *defaultHandler) deleteProvider(name string) {
	if d.registry.Handler(name) != nil {
		providerName := fmt.Sprintf("%s: %s", d.info.Factory.Name(), name)
		d.info.Handler.DeleteProvider(providerName)
		d.registry.Unregister(name)
	}
	log.Debugf("%s deleted", name)
}

func (d *defaultHandler) Count() int {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return len(d.targets)
}
