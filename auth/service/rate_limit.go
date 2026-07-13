package service

import (
	"context"
	"sync"
	"time"
)

type senderEntry struct {
	count   int
	resetAt time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*senderEntry
	stopCh  chan struct{}
}

func NewRateLimiter(ctx context.Context) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*senderEntry),
		stopCh:  make(chan struct{}),
	}
	go rl.cleanup(ctx)
	return rl
}

func (rl *RateLimiter) Allow(sender string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, ok := rl.entries[sender]
	if !ok || now.After(entry.resetAt) {
		rl.entries[sender] = &senderEntry{
			count:   1,
			resetAt: now.Add(30 * time.Second),
		}
		return true
	}

	if entry.count >= 3 {
		return false
	}

	entry.count++
	return true
}

func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

func (rl *RateLimiter) cleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rl.stopCh:
			return
		case <-ticker.C:
			rl.purgeExpired()
		}
	}
}

func (rl *RateLimiter) purgeExpired() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for sender, entry := range rl.entries {
		if now.After(entry.resetAt) {
			delete(rl.entries, sender)
		}
	}
}
