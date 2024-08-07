package internal

import (
	"fmt"
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

func TestHackAlbumYear(t *testing.T) {
	type ok struct {
		url  string
		want int
	}

	type nok struct {
		url  string
		want string
	}

	oks := []ok{
		{url: "LocalMusic:contextMenu/Song?filename=%2Fvar%2Fmnt%2FHOME-music%2F" +
			"Kamelot%2F%5B2003%5D%20Epica%2F01%20-%20Prologue.flac", want: 2003},
		{url: "LocalMusic:contextMenu/Song?filename=%2Fvar%2Fmnt%2FHOME-music%2F" +
			"Bohren%20and%20Der%20Club%20of%20Gore%2F%5B2002%5D%20Black%20Earth%2F01%20-" +
			"%20Midnight%20Black%20Earth.flac", want: 2002},
		{url: "LocalMusic:contextMenu/Song?filename=%2Fvar%2Fmnt%2FHOME-music%2F" +
			"Rage%20Against%20The%20Machine%2F%5B2000%5D%20Renegades%2F01.%20Microphone%20Fiend.flac", want: 2000},
	}

	for _, tc := range oks {
		got, _ := HackAlbumYear(tc.url)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}

	fails := []nok{
		{url: "LocalMusic:contextMenu/Song?filename=%2Fvar%2Fmnt%2FHOME-music%2F" +
			"_acoustic%2FJay%20Smith%20-%20Let%20My%20Heart%20Go.mp3", want: "year could not be found"},
	}

	for _, tc := range fails {
		_, err := HackAlbumYear(tc.url)
		if !reflect.DeepEqual(err.Error(), tc.want) {
			t.Errorf("expected: %v, got: %v", tc.want, err.Error())
		}
	}

}

func TestEscapeStyleTag(t *testing.T) {
	type test struct {
		s    string
		want string
	}

	tests := []test{
		{s: "Chilombo [clean]", want: "Chilombo [clean[]"},
		{s: "What (Deluxe)", want: "What (Deluxe)"},
		{s: "Tomorrow", want: "Tomorrow"},
	}

	for _, tc := range tests {
		got := EscapeStyleTag(tc.s)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
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

func TestCleanAlbumName(t *testing.T) {
	type test struct {
		s    string
		want string
	}

	tests := []test{
		{s: "[::b]It Was Good Until It Wasn't [clean[] (2020)", want: "It Was Good Until It Wasn't [clean]"},
	}

	for _, tc := range tests {
		got := CleanAlbumName(tc.s)
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
		fmt.Printf("Test case: %s vs %s -> Result: %.5f\n", tc.s1, tc.s2, got)

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
