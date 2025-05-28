package storage

import (
	"testing"
	"time"
)

func TestStore_SecureFlow(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	store := &Store{
		secrets: make(map[string]*Secret),
		key:     key,
	}

	text := "super secret"
	id, code, err := store.Save(text, 5, true)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	sec, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sec.Unlocked {
		t.Fatal("Secret should not be unlocked yet")
	}

	unlocked := make(chan string)
	listenerReady := make(chan struct{})

	go func() {
		listenerReady <- struct{}{} // signal: listener beginnt
		got, err := store.WaitForUnlock(id)
		if err != nil {
			t.Errorf("WaitForUnlock failed: %v", err)
			return
		}
		plain, err := store.DecryptSecretText(got)
		if err != nil {
			t.Errorf("Decrypt failed: %v", err)
			return
		}
		unlocked <- plain
	}()

	<-listenerReady

	err = store.Confirm(id, "wrong")
	if err == nil {
		t.Error("Expected Confirm to fail with wrong code")
	}

	err = store.Confirm(id, code)
	if err != nil {
		t.Fatalf("Confirm failed with correct code: %v", err)
	}

	select {
	case result := <-unlocked:
		if result != text {
			t.Errorf("Expected decrypted: %q, got %q", text, result)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for unlock result")
	}
}
