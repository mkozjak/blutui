package internal

import "testing"

func TestJWDistance(t *testing.T) {
	var d float64

	d = JWDistance("faremviel", "farmville")
	if d != 0.91898 {
		t.Error("Expected 0.91898, got ", d)
	}
}
