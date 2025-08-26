package cache

import (
	"testing"
	"time"
)

func TestTTLCache_GetSet(t *testing.T) {
	c := New[string, int](200 * time.Millisecond)
	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected miss")
	}
	c.Set("a", 42)
	if v, ok := c.Get("a"); !ok || v != 42 {
		t.Fatalf("expected 42, got %v ok=%v", v, ok)
	}
	time.Sleep(210 * time.Millisecond)
	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected expired")
	}
}
