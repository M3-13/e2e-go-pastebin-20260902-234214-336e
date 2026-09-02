package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pastebin/internal/store"
)

func doGet(s *store.Store, id string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pastes/"+id, nil)
	req.SetPathValue("id", id)
	GetHandler(s)(rec, req)
	return rec
}

func TestGetHandlerReturnsPaste(t *testing.T) {
	s := store.New()
	secs := 60
	p, err := s.Create("print(\"hi\")", "python", &secs)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	rec := doGet(s, p.ID)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got store.Paste
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("id = %q, want %q", got.ID, p.ID)
	}
	if got.Content != "print(\"hi\")" {
		t.Errorf("content = %q, want %q", got.Content, "print(\"hi\")")
	}
	if got.Language != "python" {
		t.Errorf("language = %q, want %q", got.Language, "python")
	}
	if got.CreatedAt.IsZero() {
		t.Error("created_at is zero")
	}
	if got.ExpiresAt == nil {
		t.Error("expires_at is nil, want set")
	}
}

func TestGetHandlerUnknownID(t *testing.T) {
	s := store.New()
	rec := doGet(s, "00000000000000000000000000000000")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetHandlerInvalidID(t *testing.T) {
	s := store.New()
	for _, id := range []string{"", "abc", "XYZ", "0000000000000000000000000000000"} {
		rec := doGet(s, id)
		if rec.Code != http.StatusNotFound {
			t.Errorf("id %q: status = %d, want %d", id, rec.Code, http.StatusNotFound)
		}
	}
}

func TestGetHandlerExpired(t *testing.T) {
	s := store.New()
	secs := -1
	p, err := s.Create("secret", "", &secs)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	rec := doGet(s, p.ID)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
