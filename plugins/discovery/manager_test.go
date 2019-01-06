package discovery

import (
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
)

func TestRuleDelete(t *testing.T) {
	d := manager()
	d.load(config(4))
	if len(d.rules) != 4 {
		t.Errorf("delete rule error. expected: 4 actual:%d", len(d.rules))
	}

	d.load(config(2))
	if len(d.rules) != 2 {
		t.Errorf("delete rule error. expected: 2 actual:%d", len(d.rules))
	}
}

func TestRuleAdd(t *testing.T) {
	d := manager()
	d.load(config(2))
	if len(d.rules) != 2 {
		t.Errorf("add rule error. expected: 2 actual:%d", len(d.rules))
	}
	d.load(config(4))
	if len(d.rules) != 4 {
		t.Errorf("add rule error. expected: 2 actual:%d", len(d.rules))
	}
}

func manager() *discoveryManager {
	return &discoveryManager{
		resourceLister:  discovery.NewFakeResourceLister(2),
		providerHandler: &util.DummyProviderHandler{},
		rules:           make(map[string]discovery.RuleHandler),
	}
}

func config(num int) discovery.Config {
	var rules []discovery.PrometheusConfig
	for i := 0; i < num; i++ {
		rule := discovery.PrometheusConfig{
			Name: "rule" + string(i),
		}
		rules = append(rules, rule)
	}
	return discovery.Config{
		PromConfigs: rules,
	}
}
