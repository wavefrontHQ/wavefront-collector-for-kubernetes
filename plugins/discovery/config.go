// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"io/ioutil"
	"reflect"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	discoveryAnnotation = "wavefront.com/discovery-config"
)

type configHandler struct {
	stopCh            chan struct{}
	configMapInformer cache.SharedInformer
	secretInformer    cache.SharedInformer

	mtx         sync.RWMutex
	cfg         discovery.Config            // main configuration obtained by combining wired and dynamic configuration
	wiredCfg    discovery.Config            // wired configuration
	runtimeCfgs map[string]discovery.Config // dynamic runtime configurations
	changed     bool                        // flag for tracking runtime cfg changes
}

type configResource struct {
	meta metav1.ObjectMeta
	data map[string]string
}

func newConfigHandler(kubeClient kubernetes.Interface, cfg discovery.Config) *configHandler {
	handler := &configHandler{
		cfg:         cfg,
		wiredCfg:    cfg,
		runtimeCfgs: make(map[string]discovery.Config),
	}

	ns := util.GetNamespaceName()
	if ns == "" {
		ns = readNamespaceFromFile()
		if ns == "" {
			return handler
		}
	}

	handler.configMapInformer = newConfigMapInformer(kubeClient, ns, handler)
	handler.secretInformer = newSecretInformer(kubeClient, ns, handler)
	return handler
}

func newConfigMapInformer(kubeClient kubernetes.Interface, ns string, handler *configHandler) cache.SharedInformer {
	s := kubeClient.CoreV1().ConfigMaps(ns)
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return s.List(context.Background(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return s.Watch(context.Background(), options)
		},
	}

	inf := cache.NewSharedInformer(lw, &v1.ConfigMap{}, 1*time.Hour)
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateConfigMapIfValid(obj, handler)
		},
		UpdateFunc: func(_, obj interface{}) {
			updateConfigMapIfValid(obj, handler)
		},
		DeleteFunc: func(obj interface{}) {
			deleteConfigMapIfValid(obj, handler)
		},
	})
	return inf
}

func deleteConfigMapIfValid(obj interface{}, handler *configHandler) {
	cfg, ok := obj.(*v1.ConfigMap)
	if ok {
		handler.deleted(cfg.Name)
	}
}

func updateConfigMapIfValid(obj interface{}, handler *configHandler) {
	cfg, ok := obj.(*v1.ConfigMap)
	if ok {
		handler.updated(&configResource{cfg.ObjectMeta, cfg.Data})
	}
}

func emptySecretList() *v1.SecretList {
	return &v1.SecretList{
		ListMeta: metav1.ListMeta{
			Continue:           "false",
			RemainingItemCount: nil,
			ResourceVersion:    "0",
		},
		Items: make([]v1.Secret, 0),
	}
}

func newSecretInformer(kubeClient kubernetes.Interface, ns string, handler *configHandler) cache.SharedInformer {
	authChecker := NewAuthChecker(kubeClient.AuthorizationV1().SelfSubjectAccessReviews(),
		ns, time.Minute, 12*time.Hour)
	s := kubeClient.CoreV1().Secrets(ns)
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			if authChecker.CanListSecrets() {
				return s.List(context.Background(), options)
			} else {
				return emptySecretList(), nil
			}
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return s.Watch(context.Background(), options)
		},
	}

	inf := cache.NewSharedInformer(lw, &v1.Secret{}, 1*time.Hour)
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateSecretIfValid(obj, handler)
		},
		UpdateFunc: func(_, obj interface{}) {
			updateSecretIfValid(obj, handler)
		},
		DeleteFunc: func(obj interface{}) {
			deleteSecretIfValid(obj, handler)
		},
	})
	return inf
}

func deleteSecretIfValid(obj interface{}, handler *configHandler) {
	secret, ok := obj.(*v1.Secret)
	if ok {
		handler.deleted(secret.Name)
	}
}

func updateSecretIfValid(obj interface{}, handler *configHandler) {
	secret, ok := obj.(*v1.Secret)
	if ok {
		handler.updated(&configResource{secret.ObjectMeta, convertByteArrayData(secret.Data)})
	}
}

// Config gets the combined discovery configuration and a boolean indicating whether
// the configuration has changed since the last call to this function
func (handler *configHandler) Config() (discovery.Config, bool) {
	handler.mtx.Lock()
	defer handler.mtx.Unlock()

	if !handler.changed {
		return handler.cfg, false
	}
	handler.changed = false

	newCfg := combine(handler.wiredCfg, handler.runtimeCfgs)
	if !reflect.DeepEqual(handler.cfg, newCfg) {
		// update the main combined config
		handler.cfg = newCfg
		return handler.cfg, true
	}
	return handler.cfg, false
}

func (handler *configHandler) updated(configResource *configResource) {
	if !annotated(configResource.meta.Annotations) {
		// delegate to deleted and return
		log.Infof("no runtime annotation on %s", configResource.meta.Name)
		handler.deleted(configResource.meta.Name)
		return
	}

	loaded, err := load(configResource.data)
	if err != nil {
		log.Errorf("error loading discovery config: %s error: %v", configResource.meta.Name, err)
		return
	}
	log.Infof("loaded discovery configuration from %s", configResource.meta.Name)

	handler.mtx.Lock()
	defer handler.mtx.Unlock()

	// update the internal map entry
	handler.runtimeCfgs[configResource.meta.Name] = loaded
	handler.changed = true
}

func (handler *configHandler) deleted(name string) {
	handler.mtx.Lock()
	defer handler.mtx.Unlock()
	if _, found := handler.runtimeCfgs[name]; found {
		log.Infof("deleted discovery configuration from %s", name)
		delete(handler.runtimeCfgs, name)
		handler.changed = true
	}
}

func annotated(annotations map[string]string) bool {
	if val, ok := annotations[discoveryAnnotation]; ok {
		return val == "true"
	}
	return false
}

func load(data map[string]string) (discovery.Config, error) {
	cfg := &discovery.Config{}
	for _, config := range data {
		loadedCfg, err := discovery.FromYAML([]byte(config))
		if err != nil {
			return *cfg, err
		}
		cfg.PluginConfigs = append(cfg.PluginConfigs, loadedCfg.PluginConfigs...)
	}
	return *cfg, nil
}

func combine(cfg discovery.Config, cfgs map[string]discovery.Config) discovery.Config {
	runCfg := &discovery.Config{
		DiscoveryInterval:          cfg.DiscoveryInterval,
		AnnotationPrefix:           cfg.AnnotationPrefix,
		AnnotationExcludes:         cfg.AnnotationExcludes,
		PluginConfigs:              cfg.PluginConfigs,
		DisableAnnotationDiscovery: cfg.DisableAnnotationDiscovery,
	}

	// build a sorted slice of map keys for consistent iteration order
	keys := make([]string, len(cfgs))
	i := 0
	for k := range cfgs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	log.Debug("combining discovery configurations")
	for _, key := range keys {
		c := cfgs[key]
		if len(c.PluginConfigs) > 0 {
			runCfg.PluginConfigs = append(runCfg.PluginConfigs, c.PluginConfigs...)
		}
	}
	log.Debugf("total plugin configs: %d", len(runCfg.PluginConfigs))
	return *runCfg
}

func readNamespaceFromFile() string {
	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Errorf("error reading namespace: %v", err)
		return ""
	}
	ns := string(data)
	log.Infof("loaded namespace from file: %s", ns)
	return ns
}

func convertByteArrayData(data map[string][]byte) map[string]string {
	stringData := make(map[string]string)
	for key, value := range data {
		stringData[key] = string(value)
	}
	return stringData
}

func (handler *configHandler) start() bool {
	handler.stopCh = make(chan struct{})
	go handler.configMapInformer.Run(handler.stopCh)
	go handler.secretInformer.Run(handler.stopCh)
	return cache.WaitForCacheSync(handler.stopCh, handler.configMapInformer.HasSynced, handler.secretInformer.HasSynced)
}

func (handler *configHandler) stop() {
	if handler.stopCh != nil {
		close(handler.stopCh)
	}
}
