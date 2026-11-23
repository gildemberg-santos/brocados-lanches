package ratelimit

import (
	"testing"
	"time"
)

func TestNewRejectsInvalidConfig(t *testing.T) {
	if _, err := New(0, time.Minute); err == nil {
		t.Error("New(0, minute) error = nil, want error")
	}
	if _, err := New(5, 0); err == nil {
		t.Error("New(5, 0) error = nil, want error")
	}
}

func TestAllowResetsAfterWindow(t *testing.T) {
	limiter, err := New(2, time.Minute)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	if !limiter.Allow("ip|email") || !limiter.Allow("ip|email") {
		t.Fatal("first two requests should be allowed")
	}
	if limiter.Allow("ip|email") {
		t.Error("third request should be blocked")
	}
	if !limiter.Allow("outra-chave") {
		t.Error("different key should be allowed")
	}

	now = now.Add(time.Minute)
	if !limiter.Allow("ip|email") {
		t.Error("request after window should be allowed")
	}
}

func TestPurgeRemovesExpiredKeys(t *testing.T) {
	limiter, err := New(1, time.Minute)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	for index := 0; index < maxTrackedKeys; index++ {
		limiter.Allow(string(rune(index)))
	}
	now = now.Add(time.Minute)
	if !limiter.Allow("nova-chave") {
		t.Fatal("new key should be allowed after purge")
	}
	if len(limiter.hits) > 2 {
		t.Errorf("tracked keys = %d, want at most 2 after purge", len(limiter.hits))
	}
}
