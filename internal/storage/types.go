package storage

import (
	"time"
)

type Secret struct {
	CipherText  string
	Nonce       []byte
	ExpiresAt   time.Time
	Code        string
	Unlocked    bool
	WaitingCh   chan struct{}
	listenerSet bool
}
