package httpx

import (
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
