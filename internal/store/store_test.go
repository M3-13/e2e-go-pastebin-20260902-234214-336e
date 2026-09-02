package store

import (
	"sync"
	"testing"
)

func TestCreateGeneratesID(t *testing.T) {
	s := New()
	p, err := s.Create("hello", "", nil)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if len(p.ID) != 32 {
		t.Fatalf("ID length = %d, want 32", len(p.ID))
	}
	for _, c := range p.ID {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("ID contains non-hex char %q", c)
		}
	}
	if p.CreatedAt.IsZero() {
		t.Fatal("CreatedAt is zero")
	}
}

func TestCreateUniqueIDs(t *testing.T) {
	s := New()
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		p, err := s.Create("x", "", nil)
		if err != nil {
			t.Fatalf("Create error: %v", err)
		}
		if ids[p.ID] {
			t.Fatalf("duplicate ID generated: %s", p.ID)
		}
		ids[p.ID] = true
	}
}

func TestCreateWithExpiry(t *testing.T) {
	s := New()
	secs := 60
	p, err := s.Create("hello", "go", &secs)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if p.ExpiresAt == nil {
		t.Fatal("ExpiresAt is nil, want set")
	}
	if p.ExpiresAt.Before(p.CreatedAt) {
		t.Fatal("ExpiresAt before CreatedAt")
	}
	if p.Language != "go" {
		t.Fatalf("Language = %q, want %q", p.Language, "go")
	}
}

func TestGet(t *testing.T) {
	s := New()
	p, _ := s.Create("hello", "", nil)

	got, ok := s.Get(p.ID)
	if !ok {
		t.Fatal("Get returned ok=false for existing paste")
	}
	if got.Content != "hello" {
		t.Fatalf("Content = %q, want %q", got.Content, "hello")
	}
}

func TestGetUnknown(t *testing.T) {
	s := New()
	if _, ok := s.Get("00000000000000000000000000000000"); ok {
		t.Fatal("Get returned ok=true for unknown paste")
	}
}

func TestGetExpired(t *testing.T) {
	s := New()
	secs := -1
	p, _ := s.Create("hello", "", &secs)

	if _, ok := s.Get(p.ID); ok {
		t.Fatal("Get returned ok=true for expired paste")
	}
}

func TestListOmitsExpired(t *testing.T) {
	s := New()
	p1, _ := s.Create("a", "", nil)
	secs := -1
	p2, _ := s.Create("b", "", &secs)

	list := s.List()
	ids := make(map[string]bool)
	for _, p := range list {
		ids[p.ID] = true
	}
	if !ids[p1.ID] {
		t.Fatal("active paste missing from List")
	}
	if ids[p2.ID] {
		t.Fatal("expired paste present in List")
	}
}

func TestDelete(t *testing.T) {
	s := New()
	p, _ := s.Create("hello", "", nil)

	if !s.Delete(p.ID) {
		t.Fatal("Delete returned false for existing paste")
	}
	if _, ok := s.Get(p.ID); ok {
		t.Fatal("paste still present after Delete")
	}
	if s.Delete(p.ID) {
		t.Fatal("second Delete returned true, want false")
	}
}

func TestDeleteExpired(t *testing.T) {
	s := New()
	secs := -1
	p, _ := s.Create("hello", "", &secs)

	if s.Delete(p.ID) {
		t.Fatal("Delete returned true for expired paste, want false")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := s.Create("content", "", nil)
			if err != nil {
				t.Errorf("Create error: %v", err)
			}
		}(i)
	}
	wg.Wait()

	if len(s.List()) != 50 {
		t.Fatalf("List length = %d, want 50", len(s.List()))
	}
}

func TestExpiresAtNil(t *testing.T) {
	s := New()
	p, _ := s.Create("hello", "", nil)
	if p.ExpiresAt != nil {
		t.Fatal("ExpiresAt should be nil when no expiry")
	}
}
