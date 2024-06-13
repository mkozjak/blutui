package internal

import "testing"

func TestJWSimilarity(t *testing.T) {
	var d float64

	d = JWSimilarity("faremviel", "farmville")
	if d != 0.91898 {
		t.Error("Expected 0.91898, got ", d)
	}
}
