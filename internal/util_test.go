package internal

import "testing"

func TestJWSimilarity(t *testing.T) {
	var d float64

	d = JWSimilarity("faremviel", "farmville")
	if d != 0.91898 {
		t.Error("Expected 0.91898, got ", d)
	}
}

func TestNoJWSimilarity(t *testing.T) {
	var d float64

	d = JWSimilarity("foo", "bar")
	if d != 0 {
		t.Error("Expected 0, got ", d)
	}
}

func TestExactJWSimilarity(t *testing.T) {
	var d float64

	d = JWSimilarity("london", "london")
	if d != 1 {
		t.Error("Expected 1, got ", d)
	}
}

var result float64

func BenchmarkJWSimilarity(b *testing.B) {
	var r float64

	for n := 0; n < b.N; n++ {
		r = JWSimilarity("faremviel", "farmville")
	}

	result = r
}
