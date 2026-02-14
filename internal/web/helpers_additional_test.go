package web

import (
	"errors"
	"testing"
)

func TestErrorMarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()

		b, err := (Error{}).MarshalJSON()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(b) != `{"error": null}` {
			t.Fatalf("unexpected JSON: %s", string(b))
		}
	})

	t.Run("non nil error", func(t *testing.T) {
		t.Parallel()

		b, err := (Error{Err: errors.New("boom")}).MarshalJSON()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(b) != `{"error": "boom"}` {
			t.Fatalf("unexpected JSON: %s", string(b))
		}
	})
}

func TestDetectContentType(t *testing.T) {
	t.Parallel()

	if got := detectContentType("file.JSON", []byte("not-json")); got != "application/json" {
		t.Fatalf("unexpected content type for json extension: %q", got)
	}

	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	if got := detectContentType("file.bin", pngHeader); got != "image/png" {
		t.Fatalf("unexpected content type from sniffing: %q", got)
	}
}

func TestProcessRouteForPrometheus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		uri      string
		expected string
	}{
		{uri: "/api/cache/abc/def", expected: "/api/cache"},
		{uri: "/api/cache/", expected: "/api/cache"},
		{uri: "/api/info", expected: "/api/info"},
	}

	for _, tc := range cases {
		got := processRouteForPrometheus(tc.uri)
		if got != tc.expected {
			t.Fatalf("for %q: got %q, want %q", tc.uri, got, tc.expected)
		}
	}
}
