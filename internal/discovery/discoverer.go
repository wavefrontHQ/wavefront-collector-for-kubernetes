package discovery

type defaultDiscoverer struct {
	runtimeHandler TargetHandler
}

func NewDiscoverer(targetHandler TargetHandler) Discoverer {
	return &defaultDiscoverer{
		runtimeHandler: targetHandler,
	}
}

func (d *defaultDiscoverer) Discover(resource Resource) {
	d.runtimeHandler.Handle(resource, nil)
}

func (d *defaultDiscoverer) Delete(resource Resource) {
	name := ResourceName(resource.Kind, resource.Meta)
	d.runtimeHandler.Delete(name)
}
