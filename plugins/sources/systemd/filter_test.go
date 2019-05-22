package systemd

import "testing"

func TestFromQuery(t *testing.T) {
	vals := make(map[string][]string)
	if fromQuery(vals) != nil {
		t.Errorf("error creating filter")
	}

	// test whitelisting
	vals["unitWhitelist"] = []string{"docker*", "kubelet*"}
	f := fromQuery(vals)
	if f == nil {
		t.Errorf("error creating filter")
	}
	if f.unitWhitelist == nil {
		t.Errorf("error creating filter")
	}
	if !f.match("docker.service") {
		t.Errorf("error matching whitelisted docker.service")
	}
	if !f.match("kubelet.service") {
		t.Errorf("error matching whitelisted kubelet.service")
	}
	if f.match("random.service") {
		t.Errorf("error matching random.service")
	}

	// test blacklisting
	delete(vals, "unitWhitelist")
	vals["unitBlacklist"] = []string{"*mount*", "etc*"}
	f = fromQuery(vals)
	if f.match("home.mount") {
		t.Errorf("error matching blacklisted home.mount")
	}
	if !f.match("kubelet.service") {
		t.Errorf("error matching kubelet.service")
	}
}
