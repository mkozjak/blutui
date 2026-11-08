package player

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type statusTransport func(*http.Request) (*http.Response, error)

func (f statusTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

func TestFetchStatus(t *testing.T) {
	for _, tc := range []struct {
		name       string
		code       int
		body       string
		wantError  bool
		wantStatus bool
	}{
		{"track", 200, `<status etag="next"><state>play</state><artist>Artist</artist><name>Song</name></status>`, false, true},
		{"unchanged", 304, "", false, false},
		{"server error", 503, `<status><state>play</state></status>`, true, false},
		{"malformed", 200, `<status>`, true, false},
		{"missing state", 200, `<status/>`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{Reader: strings.NewReader(tc.body)}
			client := &http.Client{Transport: statusTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Query().Get("etag") != "a&b+ c" || r.URL.Query().Get("timeout") != "60" {
					t.Errorf("incorrect status query: %s", r.URL.RawQuery)
				}

				return &http.Response{StatusCode: tc.code, Status: fmt.Sprint(tc.code), Body: body, Header: make(http.Header)}, nil
			})}
			p := New("http://player", nil, nil)
			s, err := p.fetchStatus(context.Background(), client, "a&b+ c")

			if (err != nil) != tc.wantError || (s != nil) != tc.wantStatus {
				t.Fatalf("status = %+v, error = %v", s, err)
			}

			if !body.closed {
				t.Fatal("status response body left open")
			}

			if s != nil && (s.Artist != "Artist" || s.Track != "Song") {
				t.Fatalf("incorrect metadata: %+v", s)
			}
		})
	}
}

func TestPollStatusTrackChangesAndRecovery(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	var calls atomic.Int32
	http.DefaultTransport = statusTransport(func(r *http.Request) (*http.Response, error) {
		n := calls.Add(1)

		if n > 4 {
			<-r.Context().Done()
			return nil, r.Context().Err()
		}

		wantETag := ""

		if n == 2 {
			wantETag = "one"
		}
		if n == 3 {
			wantETag = "two"
		}

		if got := r.URL.Query().Get("etag"); got != wantETag {
			t.Errorf("request %d: etag = %q, want %q", n, got, wantETag)
		}

		if n == 3 {
			return nil, io.ErrUnexpectedEOF
		}

		etag, track := "one", "First"

		if n == 2 {
			etag, track = "two", "Second"
		}
		if n == 4 {
			etag, track = "three", "Recovered"
		}

		body := fmt.Sprintf(`<status etag="%s"><state>play</state><artist>Artist</artist><name>%s</name></status>`, etag, track)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := make(chan Status)
	p := New("http://player", nil, updates)
	done := make(chan struct{})
	go func() { defer close(done); p.pollStatus(ctx) }()

	for _, want := range []string{"First", "Second", "neterr", "Recovered"} {
		select {
		case s := <-updates:
			if s.Track != want && s.State != want {
				t.Fatalf("update = %+v, want %q", s, want)
			}
			p.State()
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out waiting for %q", want)
		}
	}

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("poller did not stop")
	}
}
