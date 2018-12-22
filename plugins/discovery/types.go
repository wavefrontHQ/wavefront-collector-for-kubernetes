package discovery

type discoverer interface {
	discover(cfg Config)
}
