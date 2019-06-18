package discovery

import (
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func newPodHandler(kubeClient kubernetes.Interface, discoverer discovery.Discoverer) {
	client := kubeClient.CoreV1().RESTClient()
	fieldSelector := util.GetFieldSelector("pods")
	lw := cache.NewListWatchFromClient(client, "pods", v1.NamespaceAll, fieldSelector)
	inf := cache.NewSharedInformer(lw, &v1.Pod{}, 10*time.Minute)

	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			discoverer.Discover(discovery.Resource{
				Kind:    discovery.PodType.String(),
				IP:      pod.Status.PodIP,
				Meta:    pod.ObjectMeta,
				PodSpec: pod.Spec,
			})
		},
		UpdateFunc: func(_, obj interface{}) {
			pod := obj.(*v1.Pod)
			discoverer.Discover(discovery.Resource{
				Kind:    discovery.PodType.String(),
				IP:      pod.Status.PodIP,
				Meta:    pod.ObjectMeta,
				PodSpec: pod.Spec,
			})
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			discoverer.Delete(discovery.Resource{
				Kind:    discovery.PodType.String(),
				IP:      pod.Status.PodIP,
				Meta:    pod.ObjectMeta,
				PodSpec: pod.Spec,
			})
		},
	})
	go inf.Run(wait.NeverStop)
}
