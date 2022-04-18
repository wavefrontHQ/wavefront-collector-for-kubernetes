package discovery

type PluginConfigProvider interface {
	PluginConfigs() []PluginConfig
}
