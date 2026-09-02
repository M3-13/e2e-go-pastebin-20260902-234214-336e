package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, map[string]string{"a": "b"})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"a":"b"}` {
		t.Fatalf("body = %q, want %q", body, `{"a":"b"}`)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusNotFound, "not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"error":"not found"}` {
		t.Fatalf("body = %q, want %q", body, `{"error":"not found"}`)
	}
}

func TestIsValidID(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"0123456789abcdef0123456789abcdef", true},
		{"abcdefabcdefabcdefabcdefabcdefab", true},
		{"", false},
		{"0123456789abcdef0123456789abcde", false},
		{"0123456789abcdef0123456789abcdeg", false},
		{"0123456789ABCDEF0123456789ABCDEF", false},
		{"0123456789abcdef0123456789abcde ", false},
	}
	for _, c := range cases {
		if got := isValidID(c.id); got != c.want {
			t.Errorf("isValidID(%q) = %v, want %v", c.id, got, c.want)
		}
	}
}
