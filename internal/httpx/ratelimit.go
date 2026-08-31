// Package httpx holds small net/http middleware not tied to auth specifically.
package httpx

import (
	"net"
	"net/http"
	"strings"
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
	// lastSeen tracks the newest request per IP so stale keys can be pruned;
	// otherwise the map grows without bound over a flood of distinct IPs.
	lastSeen map[string]time.Time
}

func NewIPRateLimiter(max int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		window:   window,
		max:      max,
		requests: make(map[string][]time.Time),
		lastSeen: make(map[string]time.Time),
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
		l.record(ip, now)
		return false
	}
	l.requests[ip] = append(kept, now)
	l.record(ip, now)
	return true
}

// record updates lastSeen and prunes idle entries once the map gets large.
// Pruning only on growth keeps the steady-state cost at one map clean per
// threshold crossing instead of per request.
func (l *IPRateLimiter) record(ip string, now time.Time) {
	l.lastSeen[ip] = now
	if len(l.lastSeen) < 10000 {
		return
	}
	idle := now.Add(-2 * l.window)
	for k, t := range l.lastSeen {
		if t.Before(idle) {
			delete(l.requests, k)
			delete(l.lastSeen, k)
		}
	}
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
	// Behind the Caddy reverse proxy the remote address is the proxy itself,
	// so read the real client from X-Forwarded-For when it is set. The first
	// (leftmost) value is the client as set by Caddy; we never trust values
	// added by the client, since Caddy overwrites/appends on ingress.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
