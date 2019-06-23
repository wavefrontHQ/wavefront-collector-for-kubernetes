package telegraf

import (
	"fmt"
	"net/url"
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

func NewEncoder() discovery.Encoder {
	return telegrafEncoder{}
}

func (e telegrafEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) url.Values {
	if ip == "" {
		return url.Values{}
	}

	// panics if rule is not of expected type
	cfg := rule.(discovery.PluginConfig)
	name := discovery.ResourceName(kind, meta)
	pluginName := strings.Replace(cfg.Type, "telegraf/", "", -1)

	values := url.Values{}
	values.Set("discovered", "true")
	values.Set("plugins", pluginName)
	values.Set("name", name)

	// parse telegraf configuration
	//TODO: optimize?
	scheme := utils.Param(meta, "", cfg.Scheme, "http")
	server := fmt.Sprintf("%s://%s:%s", scheme, ip, cfg.Port)
	conf := strings.Replace(cfg.Conf, "${server}", server, -1)
	conf = strings.Replace(conf, "${host}", ip, -1)
	conf = strings.Replace(conf, "${port}", cfg.Port, -1)
	values.Set("tg.conf", conf)

	// parse prefix, tags, labels and filters
	prefix := utils.Param(meta, discovery.PrefixAnnotation, cfg.Prefix, "")
	includeLabels := utils.Param(meta, discovery.LabelsAnnotation, cfg.IncludeLabels, "true")

	values.Set("prefix", prefix)
	utils.EncodeMeta(values, kind, meta)
	utils.EncodeTags(values, "", cfg.Tags)
	if includeLabels == "true" {
		utils.EncodeTags(values, "label.", meta.Labels)
	}
	utils.EncodeFilters(values, cfg.Filters)

	return values
}
