package discovery

import "testing"

func TestUniqueness(t *testing.T) {
	// create multiple registries of same name and verify only one exists
	r1 := NewRegistry("test")
	r2 := NewRegistry("test")
	if r1 != r2 {
		t.Error("registry is not unique")
	}

	if len(registries) != 1 {
		t.Error("invalid number of registries")
	}
}
