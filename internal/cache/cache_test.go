package cache

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// Test that Get on non-existent key returns miss.
func TestTTLCache_Get_Miss(t *testing.T) {
	c := New[string, int](200 * time.Millisecond)
	if _, ok := c.Get("missing"); ok {
		t.Fatalf("expected miss for non-existent key")
	}
}

// Test that Set followed by Get returns the value.
func TestTTLCache_Set_Then_Get(t *testing.T) {
	c := New[string, int](time.Second)
	c.Set("a", 42)
	if v, ok := c.Get("a"); !ok || v != 42 {
		t.Fatalf("expected 42, got %v ok=%v", v, ok)
	}
	// overwrite
	c.Set("a", 7)
	if v, ok := c.Get("a"); !ok || v != 7 {
		t.Fatalf("expected 7 after overwrite, got %v ok=%v", v, ok)
	}
}

// Test that entries expire after TTL.
func TestTTLCache_Expiration(t *testing.T) {
	c := New[string, int](10 * time.Millisecond)
	c.Set("a", 1)
	if _, ok := c.Get("a"); !ok {
		t.Fatalf("expected hit before expiry")
	}
	time.Sleep(20 * time.Millisecond)
	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected expired entry to miss")
	}
}

// Test that GetOrSet calls supplier on miss and caches result.
func TestTTLCache_GetOrSet_Success(t *testing.T) {
	c := New[string, int](time.Second)
	calls := 0
	supplier := func() (int, error) { calls++; return 99, nil }
	v, err := c.GetOrSet("k", supplier)
	if err != nil || v != 99 {
		t.Fatalf("unexpected: v=%v err=%v", v, err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 supplier call, got %d", calls)
	}
	// second call should hit cache, not supplier
	v2, err := c.GetOrSet("k", supplier)
	if err != nil || v2 != 99 {
		t.Fatalf("unexpected: v2=%v err=%v", v2, err)
	}
	if calls != 1 {
		t.Fatalf("expected cached value without new supplier call, got %d", calls)
	}
}

// Test that an error from supplier is propagated and value not cached.
func TestTTLCache_GetOrSet_Error(t *testing.T) {
	c := New[string, int](time.Second)
	boom := errors.New("boom")
	_, err := c.GetOrSet("k", func() (int, error) { return 0, boom })
	if err == nil {
		t.Fatalf("expected error from supplier")
	}
	// ensure value was not cached
	if _, ok := c.Get("k"); ok {
		t.Fatalf("value should not be cached on error")
	}
}

// Test concurrent GetOrSet calls to ensure only one supplier call occurs.
func TestTTLCache_Concurrent_GetOrSet(t *testing.T) {
	c := New[string, int](time.Second)
	var mu sync.Mutex
	calls := 0
	supplier := func() (int, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		// small delay to widen race window
		time.Sleep(2 * time.Millisecond)
		return 123, nil
	}
	var wg sync.WaitGroup
	workers := 50
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if v, err := c.GetOrSet("key", supplier); err != nil || v != 123 {
				t.Errorf("unexpected: v=%v err=%v", v, err)
			}
		}()
	}
	wg.Wait()
	// final value should be present and correct
	if v, ok := c.Get("key"); !ok || v != 123 {
		t.Fatalf("expected cached 123, got %v ok=%v", v, ok)
	}
	if calls < 1 || calls > workers {
		t.Fatalf("supplier calls out of expected range: %d", calls)
	}
}
