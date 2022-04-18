package discovery

type ConfigProvider interface {
	DiscoveryConfigs() []PluginConfig
}
