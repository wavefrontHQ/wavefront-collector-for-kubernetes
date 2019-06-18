package redis

import (
	"fmt"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/telegraf"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var registry discovery.TargetRegistry

func init() {
	registry = discovery.NewRegistry("redis")
}

func NewTargetHandler(handler metrics.ProviderHandler) discovery.TargetHandler {
	return discovery.NewHandler(
		discovery.ProviderInfo{
			Handler: handler,
			Factory: telegraf.NewFactory(),
			Encoder: redisEncoder{},
		},
		registry,
	)
}

type redisEncoder struct{}

func (e redisEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) string {
	//TODO: rule can be nil
	return telegrafURL(ip, kind, meta)
}

func telegrafURL(ip, kind string, meta metav1.ObjectMeta) string {
	if ip == "" {
		return ""
	}
	name := discovery.ResourceName(kind, meta)
	prefix := utils.Param(meta, discovery.PrefixAnnotation, "", "")
	u := fmt.Sprintf("?prefix=%s&plugins=redis&server=tcp://%s:6379&name=%s", prefix, ip, name)
	u = utils.EncodeMeta(u, kind, meta)
	u = utils.EncodeTags(u, "label.", meta.Labels)
	return u
}
