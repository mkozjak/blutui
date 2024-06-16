package internal

import (
	"reflect"
	"testing"
)

func TestJWSimilarity(t *testing.T) {
	type test struct {
		s1   string
		s2   string
		want float64
	}

	tests := []test{
		{s1: "faremviel", s2: "farmville", want: 0.91898},
		{s1: "korn", s2: "cynic", want: 0},
		{s1: "limp bizkit", s2: "limp bizkit", want: 1},
		{s1: "drmea theter", s2: "dream theater", want: 0.9453},
	}

	for _, tc := range tests {
		got := JWSimilarity(tc.s1, tc.s2)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

var result float64

func BenchmarkJWSimilarity(b *testing.B) {
	type benchmark struct {
		name string
		s1   string
		s2   string
	}

	benchmarks := []benchmark{
		{s1: "faremviel", s2: "farmville"},
		{s1: "foo", s2: "bar"},
		{s1: "london", s2: "london"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				JWSimilarity(bm.s1, bm.s2)
			}
		})
	}
}
