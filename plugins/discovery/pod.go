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

func NewPodHandler(kubeClient kubernetes.Interface, discoverer discovery.Discoverer) {
	p := kubeClient.CoreV1().Pods(v1.NamespaceAll)
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return p.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return p.Watch(options)
		},
	}
	inf := cache.NewSharedInformer(lw, &v1.Service{}, 10*time.Minute)

	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			discoverer.Discover(pod.Status.PodIP, discovery.PodType.String(), pod.ObjectMeta)
		},
		UpdateFunc: func(_, obj interface{}) {
			pod := obj.(*v1.Pod)
			discoverer.Discover(pod.Status.PodIP, discovery.PodType.String(), pod.ObjectMeta)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Service)
			discoverer.Delete(discovery.PodType.String(), pod.ObjectMeta)
		},
	})
	go inf.Run(wait.NeverStop)
}
