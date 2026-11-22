// Package ratelimit implementa um limite fixo em memória para a instância atual.
// Em deployment com múltiplas réplicas, o limite deve passar para um store compartilhado.
package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

const maxTrackedKeys = 1024

type window struct {
	start time.Time
	count int
}

// Limiter conta requisições por chave em uma janela fixa.
type Limiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	hits   map[string]window
}

// New valida a configuração e cria um limitador vazio.
func New(limit int, duration time.Duration) (*Limiter, error) {
	if limit < 1 {
		return nil, fmt.Errorf("rate limit must be at least 1")
	}
	if duration <= 0 {
		return nil, fmt.Errorf("rate limit window must be positive")
	}
	return &Limiter{
		limit:  limit,
		window: duration,
		now:    time.Now,
		hits:   make(map[string]window),
	}, nil
}

// Allow informa se a chave ainda tem requisições disponíveis na janela.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	current, found := l.hits[key]
	if !found || now.Sub(current.start) >= l.window {
		if !found && len(l.hits) >= maxTrackedKeys {
			l.purge(now)
		}
		l.hits[key] = window{start: now, count: 1}
		return true
	}
	if current.count >= l.limit {
		return false
	}
	current.count++
	l.hits[key] = current
	return true
}

// purge remove janelas vencidas para não reter chaves de-identificáveis por tempo indefinido.
func (l *Limiter) purge(now time.Time) {
	for key, current := range l.hits {
		if now.Sub(current.start) >= l.window {
			delete(l.hits, key)
		}
	}
}
