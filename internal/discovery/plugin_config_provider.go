package discovery

type PluginProvider interface {
	DiscoveryPluginConfigs() []PluginConfig
}
