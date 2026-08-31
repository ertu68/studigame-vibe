package httpx

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIPRateLimiterAllowsUpToMaxThenBlocks(t *testing.T) {
	l := NewIPRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed within limit", i)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatalf("4th request should be blocked")
	}
	if !l.Allow("5.6.7.8") {
		t.Fatalf("different IP should have its own budget")
	}
}

func TestIPRateLimiterWindowExpires(t *testing.T) {
	l := NewIPRateLimiter(1, 20*time.Millisecond)

	if !l.Allow("1.2.3.4") {
		t.Fatalf("first request should be allowed")
	}
	if l.Allow("1.2.3.4") {
		t.Fatalf("second immediate request should be blocked")
	}
	time.Sleep(30 * time.Millisecond)
	if !l.Allow("1.2.3.4") {
		t.Fatalf("request after window expiry should be allowed")
	}
}

func TestClientIPReadsForwardedFor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.5:1234"
	r.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.5")
	if got := ClientIP(r); got != "203.0.113.9" {
		t.Fatalf("X-Forwarded-For: got %q, want 203.0.113.9", got)
	}

	r.Header.Del("X-Forwarded-For")
	if got := ClientIP(r); got != "10.0.0.5" {
		t.Fatalf("no XFF: got %q, want 10.0.0.5", got)
	}
}

func TestIPRateLimiterPrunesIdleIPs(t *testing.T) {
	l := NewIPRateLimiter(3, time.Minute)
	// Bypass the 10k default threshold so the prune is reachable in a test.
	l.lastSeen = make(map[string]time.Time)
	// Simulate a large map full of idle IPs that share one active peer.
	l.requests = make(map[string][]time.Time)
	for i := 0; i < 20000; i++ {
		ip := fmt.Sprintf("idle-%d", i)
		l.requests[ip] = []time.Time{time.Now().Add(-10 * time.Minute)}
		l.lastSeen[ip] = time.Now().Add(-10 * time.Minute)
	}

	// One active request crosses the prune threshold.
	if !l.Allow("active-peer") {
		t.Fatalf("active peer should be allowed")
	}
	if len(l.requests) > 100 {
		t.Fatalf("prune did not drop idle IPs, map has %d entries", len(l.requests))
	}
	if _, ok := l.requests["active-peer"]; !ok {
		t.Fatal("active peer must survive the prune")
	}
}
