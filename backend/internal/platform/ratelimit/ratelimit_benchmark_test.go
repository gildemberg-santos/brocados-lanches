package ratelimit

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func benchmarkLimiter(b *testing.B, keys int) {
	limiter, err := New(1000000, time.Minute)
	if err != nil {
		b.Fatal(err)
	}
	names := make([]string, keys)
	for index := range names {
		names[index] = fmt.Sprintf("ip|10.0.0.%d", index)
	}

	var allowed atomic.Int64
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var index atomic.Uint64
		for pb.Next() {
			key := names[(index.Add(1)-1)%uint64(len(names))]
			if limiter.Allow(key) {
				allowed.Add(1)
			}
		}
	})
	b.StopTimer()

	if allowed.Load() == 0 {
		b.Fatal("nenhuma requisição autorizada")
	}
}

// BenchmarkLimiterAllow mede Allow com 1 chave (cenário de NAT/Wi-Fi) e com
// 1024 chaves distintas, incluindo a contenção do mutex.
func BenchmarkLimiterAllow(b *testing.B) {
	for _, keys := range []int{1, 1024} {
		b.Run(fmt.Sprintf("chaves=%d", keys), func(b *testing.B) {
			benchmarkLimiter(b, keys)
		})
	}
}
