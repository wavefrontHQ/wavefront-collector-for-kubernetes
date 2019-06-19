package memcached

import (
	"fmt"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewEncoder() discovery.Encoder {
	return memcachedEncoder{}
}

type memcachedEncoder struct{}

func (e memcachedEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) string {
	//TODO: include port
	return fmt.Sprintf("plugins=memcached&server=%s:11211", ip)
}
