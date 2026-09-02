package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pastebin/internal/store"
)

func TestListHandlerMetadata(t *testing.T) {
	s := store.New()
	secs := 60
	p, err := s.Create("secret content", "go", &secs)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/pastes", nil)
	rec := httptest.NewRecorder()
	ListHandler(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var out []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1", len(out))
	}

	item := out[0]
	if item["id"] != p.ID {
		t.Fatalf("id = %v, want %v", item["id"], p.ID)
	}
	if item["language"] != "go" {
		t.Fatalf("language = %v, want go", item["language"])
	}
	if _, ok := item["created_at"]; !ok {
		t.Fatal("created_at missing from metadata")
	}
	if _, ok := item["expires_at"]; !ok {
		t.Fatal("expires_at missing from metadata")
	}
	if _, ok := item["content"]; ok {
		t.Fatal("content field must not be present in list response")
	}
}

func TestListHandlerOmitsExpired(t *testing.T) {
	s := store.New()
	active, _ := s.Create("active", "", nil)
	secs := -1
	expired, _ := s.Create("expired", "", &secs)

	req := httptest.NewRequest(http.MethodGet, "/pastes", nil)
	rec := httptest.NewRecorder()
	ListHandler(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var out []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	ids := make(map[string]bool)
	for _, item := range out {
		ids[item["id"].(string)] = true
	}
	if !ids[active.ID] {
		t.Fatal("active paste missing from list")
	}
	if ids[expired.ID] {
		t.Fatal("expired paste present in list")
	}
}

func TestListHandlerEmpty(t *testing.T) {
	s := store.New()

	req := httptest.NewRequest(http.MethodGet, "/pastes", nil)
	rec := httptest.NewRecorder()
	ListHandler(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var out []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if out == nil {
		t.Fatal("empty list must serialize to [], got null")
	}
	if len(out) != 0 {
		t.Fatalf("len = %d, want 0", len(out))
	}
}
