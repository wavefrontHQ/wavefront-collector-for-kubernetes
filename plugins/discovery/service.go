package discovery

import (
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewServiceHandler(kubeClient kubernetes.Interface, discoverer discovery.Discoverer) {
	s := kubeClient.CoreV1().Services(v1.NamespaceAll)
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return s.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return s.Watch(options)
		},
	}
	inf := cache.NewSharedInformer(lw, &v1.Service{}, 10*time.Minute)

	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			service := obj.(*v1.Service)
			discoverer.Discover(service.Spec.ClusterIP, discovery.ServiceType.String(), service.ObjectMeta)
		},
		UpdateFunc: func(_, obj interface{}) {
			service := obj.(*v1.Service)
			discoverer.Discover(service.Spec.ClusterIP, discovery.ServiceType.String(), service.ObjectMeta)
		},
		DeleteFunc: func(obj interface{}) {
			service := obj.(*v1.Service)
			discoverer.Delete(discovery.ServiceType.String(), service.ObjectMeta)
		},
	})
	go inf.Run(wait.NeverStop)
}
