package telegraf

import (
	"fmt"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/telegraf"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewTargetHandler(handler metrics.ProviderHandler, delegate discovery.Encoder, registry discovery.TargetRegistry) discovery.TargetHandler {
	return discovery.NewHandler(
		discovery.ProviderInfo{
			Handler: handler,
			Factory: telegraf.NewFactory(),
			Encoder: telegrafEncoder{delegate: delegate},
		},
		registry,
	)
}

type telegrafEncoder struct {
	delegate discovery.Encoder
}

func (e telegrafEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) string {
	//TODO: rule can be nil
	if ip == "" {
		return ""
	}
	name := discovery.ResourceName(kind, meta)
	prefix := utils.Param(meta, discovery.PrefixAnnotation, "", "")
	pluginConf := e.delegate.Encode(ip, kind, meta, rule)
	u := fmt.Sprintf("?prefix=%s&name=%s&%s", prefix, name, pluginConf)
	u = utils.EncodeMeta(u, kind, meta)
	u = utils.EncodeTags(u, "label.", meta.Labels)
	return u
}
