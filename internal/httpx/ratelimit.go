// Package httpx holds small net/http middleware not tied to auth specifically.
package httpx

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// IPRateLimiter caps requests per IP in a fixed window. It is process-local,
// in-memory state — fine for a single VPS instance.
// ponytail: single-instance only; move to a shared store (Redis) if this ever runs multi-instance.
type IPRateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	max      int
	requests map[string][]time.Time
}

func NewIPRateLimiter(max int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		window:   window,
		max:      max,
		requests: make(map[string][]time.Time),
	}
}

func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)
	kept := l.requests[ip][:0]
	for _, t := range l.requests[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.requests[ip] = kept
		return false
	}
	l.requests[ip] = append(kept, now)
	return true
}

func (l *IPRateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := ClientIP(r)
		if !l.Allow(ip) {
			http.Error(w, "zu viele Versuche, bitte spaeter erneut versuchen", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
