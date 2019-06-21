package telegraf

import (
	"fmt"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/telegraf"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultEncoder = telegrafEncoder{}

func newTargetHandler(handler metrics.ProviderHandler, plugin string) discovery.TargetHandler {
	registryName := strings.Replace(plugin, "/", ".", -1)
	return discovery.NewHandler(
		discovery.ProviderInfo{
			Handler: handler,
			Factory: telegraf.NewFactory(),
			Encoder: defaultEncoder,
		},
		discovery.NewRegistry(registryName),
	)
}

type telegrafEncoder struct{}

func (e telegrafEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) string {
	if ip == "" {
		return ""
	}
	name := discovery.ResourceName(kind, meta)
	prefix := utils.Param(meta, discovery.PrefixAnnotation, "", "")

	cfg := rule.(discovery.PluginConfig)
	scheme := utils.Param(meta, "", cfg.Scheme, "http")
	server := fmt.Sprintf("%s://%s:%s", scheme, ip, cfg.Port)

	telegrafCfg := ""
	for k, v := range cfg.Conf {
		//TODO: optimize
		v = strings.Replace(v, "${server}", server, -1)
		v = strings.Replace(v, "${host}", ip, -1)
		v = strings.Replace(v, "${port}", cfg.Port, -1)
		telegrafCfg = fmt.Sprintf("%s=%s", k, v)
	}
	pluginName := strings.Replace(cfg.Type, "telegraf/", "", -1)
	pluginConf := fmt.Sprintf("plugins=%s&%s", pluginName, telegrafCfg)

	//TODO: include prefix, tags and filters into the encoding

	u := fmt.Sprintf("?prefix=%s&name=%s&%s", prefix, name, pluginConf)
	u = utils.EncodeMeta(u, kind, meta)
	u = utils.EncodeTags(u, "label.", meta.Labels)
	return u
}
