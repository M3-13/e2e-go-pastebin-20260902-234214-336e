package store

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Paste struct {
	ID        string     `json:"id"`
	Content   string     `json:"content,omitempty"`
	Language  string     `json:"language,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type Store struct {
	mu     sync.Mutex
	pastes map[string]Paste
}

func New() *Store {
	return &Store{
		pastes: make(map[string]Paste),
	}
}

func (s *Store) Create(content, language string, expiresInSeconds *int) (Paste, error) {
	id, err := newID()
	if err != nil {
		return Paste{}, err
	}

	now := time.Now().UTC()
	p := Paste{
		ID:        id,
		Content:   content,
		Language:  language,
		CreatedAt: now,
	}

	if expiresInSeconds != nil {
		exp := now.Add(time.Duration(*expiresInSeconds) * time.Second)
		p.ExpiresAt = &exp
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pastes[id] = p
	return p, nil
}

func (s *Store) Get(id string) (Paste, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.pastes[id]
	if !ok {
		return Paste{}, false
	}
	if expired(p) {
		delete(s.pastes, id)
		return Paste{}, false
	}
	return p, true
}

func (s *Store) List() []Paste {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Paste, 0, len(s.pastes))
	now := time.Now().UTC()
	for id, p := range s.pastes {
		if expiredAt(p.ExpiresAt, now) {
			delete(s.pastes, id)
			continue
		}
		out = append(out, p)
	}
	return out
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.pastes[id]
	if !ok {
		return false
	}
	if expired(p) {
		delete(s.pastes, id)
		return false
	}
	delete(s.pastes, id)
	return true
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func expired(p Paste) bool {
	return expiredAt(p.ExpiresAt, time.Now().UTC())
}

func expiredAt(exp *time.Time, now time.Time) bool {
	return exp != nil && !exp.After(now)
}
