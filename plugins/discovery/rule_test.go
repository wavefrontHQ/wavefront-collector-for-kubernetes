package discovery

import (
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
)

func TestRuleDelete(t *testing.T) {
	rh := makeRuleHandler()

	_ = rh.HandleAll(config(4))
	if rh.Count() != 4 {
		t.Errorf("delete rule error. expected: 4 actual: %d", rh.Count())
	}

	_ = rh.HandleAll(config(2))
	if rh.Count() != 2 {
		t.Errorf("delete rule error. expected: 2 actual: %d", rh.Count())
	}
}

func TestRuleAdd(t *testing.T) {
	d := makeRuleHandler()
	_ = d.HandleAll(config(2))
	if d.Count() != 2 {
		t.Errorf("add rule error. expected: 2 actual: %d", d.Count())
	}
	_ = d.HandleAll(config(4))
	if d.Count() != 4 {
		t.Errorf("add rule error. expected: 2 actual: %d", d.Count())
	}
}

func makeRuleHandler() discovery.RuleHandler {
	return newRuleHandler(newDiscoverer(nil), true)
}

func config(num int) []discovery.PluginConfig {
	var rules []discovery.PluginConfig
	for i := 0; i < num; i++ {
		rule := discovery.PluginConfig{
			Name: "rule" + string(i),
			Type: "prometheus",
			Selectors: discovery.Selectors{
				Namespaces: []string{"default"},
			},
		}
		rules = append(rules, rule)
	}
	return rules
}
