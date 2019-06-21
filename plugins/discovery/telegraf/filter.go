package telegraf

import (
	"strconv"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"

	"github.com/gobwas/glob"
)

type resourceFilter struct {
	imageGlob glob.Glob
	port      int64
}

func newResourceFilter(conf discovery.PluginConfig) (*resourceFilter, error) {
	rf := &resourceFilter{
		imageGlob: filter.Compile(conf.Images),
	}

	// port
	val, err := strconv.ParseInt(conf.Port, 10, 32)
	if err != nil {
		return nil, err
	}
	rf.port = val

	return rf, nil
}

func (r *resourceFilter) matches(resource discovery.Resource) bool {
	for _, container := range resource.PodSpec.Containers {
		if r.imageGlob.Match(container.Image) {
			// image matches, verify matching port exists.
			for _, cPort := range container.Ports {
				if int64(cPort.ContainerPort) == r.port {
					return true
				}
			}
		}
	}
	return false
}
