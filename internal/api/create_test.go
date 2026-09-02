package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pastebin/internal/store"
)

func newTestServer() *httptest.Server {
	s := store.New()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /pastes", CreateHandler(s))
	return httptest.NewServer(mux)
}

func post(t *testing.T, srv *httptest.Server, contentType, body string) *http.Response {
	t.Helper()
	var rd *strings.Reader
	if body == "" {
		rd = strings.NewReader("")
	} else {
		rd = strings.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/pastes", rd)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	return resp
}

func TestCreateValid(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "application/json", `{"content":"hello world"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var p struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(p.ID) != 32 {
		t.Fatalf("id length = %d, want 32 (%q)", len(p.ID), p.ID)
	}
	if !isValidID(p.ID) {
		t.Fatalf("id %q not valid hex", p.ID)
	}
	if p.Content != "hello world" {
		t.Fatalf("content = %q, want %q", p.Content, "hello world")
	}
}

func TestCreateWithLanguageAndExpiry(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "application/json", `{"content":"code","language":"go","expires_in_seconds":60}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var p struct {
		Language  string `json:"language"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Language != "go" {
		t.Fatalf("language = %q, want %q", p.Language, "go")
	}
	if p.ExpiresAt == "" {
		t.Fatal("expires_at empty, want set")
	}
}

func TestCreateMissingContent(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "application/json", `{"language":"go"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCreateEmptyContent(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "application/json", `{"content":""}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCreateInvalidJSON(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "application/json", `{"content":`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCreateNonStringContent(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "application/json", `{"content":123}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCreateNegativeExpiry(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "application/json", `{"content":"hello","expires_in_seconds":-1}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCreateBodyTooLarge(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	big := strings.Repeat("a", maxBodySize+1)
	body := `{"content":"` + big + `"}`
	resp := post(t, srv, "application/json", body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

func TestCreateMissingContentType(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "", `{"content":"hello"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnsupportedMediaType)
	}
}

func TestCreateWrongContentType(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "text/plain", `{"content":"hello"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnsupportedMediaType)
	}
}

func TestCreateContentTypeWithCharset(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp := post(t, srv, "application/json; charset=utf-8", `{"content":"hello"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}
