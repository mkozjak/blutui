package internal

import (
	"reflect"
	"testing"
)

func TestFormatDuration(t *testing.T) {
	type test struct {
		sec  int
		want string
	}

	tests := []test{
		{sec: 120, want: "02:00"},
		{sec: 34, want: "00:34"},
		{sec: 0, want: "00:00"},
	}

	for _, tc := range tests {
		got := FormatDuration(tc.sec)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestCapitalizeArtist(t *testing.T) {
	type test struct {
		s    string
		want string
	}

	tests := []test{
		{s: "koRn", want: "KoRn"},
		{s: "", want: ""},
		{s: "camel", want: "Camel"},
		{s: "(foo", want: "(foo"},
	}

	for _, tc := range tests {
		got := CapitalizeArtist(tc.s)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestCaser(t *testing.T) {
	type test struct {
		s    string
		want string
	}

	tests := []test{
		{s: "koRn", want: "Korn"},
		{s: "croatia", want: "Croatia"},
		{s: "mercedes", want: "Mercedes"},
	}

	for _, tc := range tests {
		got := Caser(tc.s)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestExtractAlbumYear(t *testing.T) {
	type ok struct {
		y    string
		want int
	}

	type nok struct {
		y    string
		want string
	}

	oks := []ok{
		{y: "2023", want: 2023},
		{y: "2023-12-31", want: 2023},
		{y: "0", want: 0},
	}

	for _, tc := range oks {
		got, _ := ExtractAlbumYear(tc.y)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}

	fails := []nok{
		{y: "2x23", want: "invalid date format"},
		{y: "2000-", want: "invalid date format"},
	}

	for _, tc := range fails {
		_, err := ExtractAlbumYear(tc.y)
		if !reflect.DeepEqual(err.Error(), tc.want) {
			t.Errorf("expected: %v, got: %v", tc.want, err.Error())
		}
	}

}

func TestCleanTrackName(t *testing.T) {
	type test struct {
		s    string
		want string
	}

	tests := []test{
		{s: "01. What Is This", want: "What Is This"},
		{s: "Make Me Bad", want: "Make Me Bad"},
		{s: "mercedes", want: "mercedes"},
	}

	for _, tc := range tests {
		got := CleanTrackName(tc.s)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

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
		{s1: "inter", s2: "In", want: 0.84},
		{s1: "", s2: "In", want: 0},
		{s1: "", s2: "", want: 0},
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
