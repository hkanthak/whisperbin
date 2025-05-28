package storage

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"sync"
	"time"
)

type Store struct {
	mu      sync.Mutex
	secrets map[string]*Secret
	key     []byte
}

func NewStore() *Store {
	var key []byte
	envKey := os.Getenv("SECRET_KEY")

	if envKey != "" {
		decoded, err := base64.StdEncoding.DecodeString(envKey)
		if err != nil || len(decoded) != 32 {
			panic("invalid SECRET_KEY: must be 32-byte base64-encoded")
		}
		key = decoded
	} else {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			panic("could not generate encryption key")
		}
	}
	

	return &Store{
		secrets: make(map[string]*Secret),
		key:     key,
	}
}

func (s *Store) Save(text string, ttlMinutes int, withApproval bool) (string, string, error) {
	id, err := generateID()
	if err != nil {
		return "", "", err
	}

	cipherText, nonce, err := encrypt([]byte(text), s.key)
	if err != nil {
		return "", "", err
	}

	expiration := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
	secret := &Secret{
		CipherText: base64.StdEncoding.EncodeToString(cipherText),
		Nonce:      nonce,
		ExpiresAt:  expiration,
	}

	if withApproval {
		secret.Code = generateCode()
		secret.Unlocked = false
		secret.WaitingCh = make(chan struct{})
	} else {
		secret.Unlocked = true
	}

	s.mu.Lock()
	s.secrets[id] = secret
	s.mu.Unlock()

	return id, secret.Code, nil
}

func (s *Store) Get(id string) (*Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	secret, ok := s.secrets[id]
	if !ok || time.Now().After(secret.ExpiresAt) {
		return nil, errors.New("not found or expired")
	}
	return secret, nil
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	delete(s.secrets, id)
	s.mu.Unlock()
}

func (s *Store) Confirm(id, inputCode string) error {
	s.mu.Lock()
	sec, ok := s.secrets[id]
	if !ok || time.Now().After(sec.ExpiresAt) {
		s.mu.Unlock()
		return errors.New("not found or expired")
	}
	if sec.Code != inputCode {
		s.mu.Unlock()
		return errors.New("invalid code")
	}
	if sec.WaitingCh == nil || !sec.listenerSet {
		s.mu.Unlock()
		return errors.New("no recipient waiting")
	}
	if sec.Unlocked {
		s.mu.Unlock()
		return errors.New("already unlocked")
	}
	sec.Unlocked = true
	close(sec.WaitingCh)
	sec.WaitingCh = nil
	s.mu.Unlock()
	return nil
}

func (s *Store) WaitForUnlock(id string) (*Secret, error) {
	s.mu.Lock()
	sec, ok := s.secrets[id]
	if !ok || time.Now().After(sec.ExpiresAt) {
		s.mu.Unlock()
		return nil, errors.New("not found or expired")
	}
	if sec.WaitingCh == nil {
		s.mu.Unlock()
		return nil, errors.New("not secure mode")
	}
	if sec.listenerSet {
		s.mu.Unlock()
		return nil, errors.New("listener already connected")
	}
	sec.listenerSet = true
	ch := sec.WaitingCh
	s.mu.Unlock()

	<-ch
	return sec, nil
}

func (s *Store) DecryptSecretText(sec *Secret) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(sec.CipherText)
	if err != nil {
		return "", err
	}
	plain, err := decrypt(cipherBytes, sec.Nonce, s.key)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (s *Store) IsWaiting(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	secret, ok := s.secrets[id]
	if !ok || time.Now().After(secret.ExpiresAt) {
		return false, errors.New("not found or expired")
	}
	waiting := secret.WaitingCh != nil && !secret.Unlocked
	return waiting, nil
}
