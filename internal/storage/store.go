package storage

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

type Secret struct {
	Text      string
	ExpiresAt time.Time
}

type Store struct {
	mu      sync.Mutex
	secrets map[string]Secret
}

func NewStore() *Store {
	return &Store{
		secrets: make(map[string]Secret),
	}
}

func (s *Store) Save(text string, ttlMinutes int) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}

	expiration := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[id] = Secret{Text: text, ExpiresAt: expiration}
	return id, nil
}

func (s *Store) LoadAndDelete(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	secret, ok := s.secrets[id]
	if !ok {
		return "", errors.New("not found or already accessed")
	}
	if time.Now().After(secret.ExpiresAt) {
		delete(s.secrets, id)
		return "", errors.New("expired")
	}

	delete(s.secrets, id)
	return secret.Text, nil
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
