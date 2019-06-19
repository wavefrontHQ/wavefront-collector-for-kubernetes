package redis

import (
	"fmt"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewEncoder() discovery.Encoder {
	return redisEncoder{}
}

type redisEncoder struct{}

func (e redisEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) string {
	//TODO: include port
	return fmt.Sprintf("plugins=redis&server=tcp://%s:6379", ip)
}
