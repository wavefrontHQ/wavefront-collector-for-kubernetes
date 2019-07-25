package discovery

import (
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ResourceName(kind string, meta metav1.ObjectMeta) string {
	if kind == ServiceType.String() {
		return meta.Namespace + "-" + kind + "-" + meta.Name
	}
	return kind + "-" + meta.Name
}

// converts deprecated prometheus configs to plugin configs
func ConvertPromToPlugin(cfg *Config) {
	// convert PrometheusConfigs to PluginConfigs
	if len(cfg.PromConfigs) > 0 {
		log.Warningf("Warning: PrometheusConfig has been deprecated. Use PluginConfig.")
		toAppend := make([]PluginConfig, len(cfg.PromConfigs))
		for i, promCfg := range cfg.PromConfigs {
			toAppend[i] = PluginConfig{
				Name:          promCfg.Name,
				Type:          "prometheus",
				Port:          promCfg.Port,
				Scheme:        promCfg.Scheme,
				Path:          promCfg.Path,
				Source:        promCfg.Source,
				Prefix:        promCfg.Prefix,
				Tags:          promCfg.Tags,
				IncludeLabels: promCfg.IncludeLabels,
				Filters:       promCfg.Filters,
				Selectors: Selectors{
					ResourceType: promCfg.ResourceType,
				},
			}

			if len(promCfg.Namespace) > 0 {
				toAppend[i].Selectors.Namespaces = []string{promCfg.Namespace}
			}

			if len(promCfg.Labels) > 0 {
				labels := map[string][]string{}
				for k, v := range promCfg.Labels {
					labels[k] = []string{v}
				}
				toAppend[i].Selectors.Labels = labels
			}
		}
		cfg.PluginConfigs = append(cfg.PluginConfigs, toAppend...)
	}
}
