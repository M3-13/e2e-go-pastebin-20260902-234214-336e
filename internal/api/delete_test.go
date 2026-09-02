package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pastebin/internal/store"
)

func TestDeleteHandler(t *testing.T) {
	s := store.New()
	p, err := s.Create("hello", "", nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	handler := DeleteHandler(s)

	req := httptest.NewRequest(http.MethodDelete, "/pastes/"+p.ID, nil)
	req.SetPathValue("id", p.ID)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("first delete status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "" {
		t.Fatalf("first delete body = %q, want empty", body)
	}

	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("second delete status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"error":"not found"}` {
		t.Fatalf("second delete body = %q, want %q", body, `{"error":"not found"}`)
	}
}

func TestDeleteHandlerUnknownID(t *testing.T) {
	s := store.New()
	handler := DeleteHandler(s)

	req := httptest.NewRequest(http.MethodDelete, "/pastes/0123456789abcdef0123456789abcdef", nil)
	req.SetPathValue("id", "0123456789abcdef0123456789abcdef")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"error":"not found"}` {
		t.Fatalf("body = %q, want %q", body, `{"error":"not found"}`)
	}
}

func TestDeleteHandlerMalformedID(t *testing.T) {
	s := store.New()
	handler := DeleteHandler(s)

	for _, id := range []string{"", "short", "0123456789ABCDEF0123456789ABCDEF", "0123456789abcdef0123456789abcde"} {
		req := httptest.NewRequest(http.MethodDelete, "/pastes/"+id, nil)
		req.SetPathValue("id", id)
		rec := httptest.NewRecorder()
		handler(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("id %q: status = %d, want %d", id, rec.Code, http.StatusNotFound)
		}
	}
}
