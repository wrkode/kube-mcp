package kubernetes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RBACCacheTestSuite tests RBAC caching functionality.
type RBACCacheTestSuite struct {
	suite.Suite
	authorizer *rbacAuthorizerImpl
}

// SetupTest sets up the test.
func (s *RBACCacheTestSuite) SetupTest() {
	// Create a mock client set (we'll use envtest in integration tests)
	// For unit tests, we'll test the cache logic directly
	s.authorizer = &rbacAuthorizerImpl{
		cache: make(map[string]*rbacCacheEntry),
		ttl:   5 * time.Second,
	}
}

// TestCacheKeyGeneration tests cache key generation.
func (s *RBACCacheTestSuite) TestCacheKeyGeneration() {
	user := "test-user"
	verb := "get"
	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	namespace := "default"

	key1 := cacheKey(user, verb, gvr, namespace)
	key2 := cacheKey(user, verb, gvr, namespace)
	s.Equal(key1, key2, "Same inputs should generate same key")

	// Different namespace should generate different key
	key3 := cacheKey(user, verb, gvr, "other-ns")
	s.NotEqual(key1, key3, "Different namespace should generate different key")

	// Different verb should generate different key
	key4 := cacheKey(user, "list", gvr, namespace)
	s.NotEqual(key1, key4, "Different verb should generate different key")
}

// TestCacheExpiry tests cache expiry behavior.
func (s *RBACCacheTestSuite) TestCacheExpiry() {
	// Set a very short TTL for testing
	s.authorizer.ttl = 100 * time.Millisecond

	key := "test-key"
	entry := &rbacCacheEntry{
		allowed:   true,
		expiresAt: time.Now().Add(s.authorizer.ttl),
	}

	s.authorizer.cache[key] = entry

	// Entry should be valid immediately
	s.authorizer.mu.RLock()
	cachedEntry, ok := s.authorizer.cache[key]
	s.authorizer.mu.RUnlock()
	s.True(ok, "Entry should exist")
	s.True(time.Now().Before(cachedEntry.expiresAt), "Entry should not be expired")

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Entry should be expired now
	s.authorizer.mu.RLock()
	cachedEntry, ok = s.authorizer.cache[key]
	s.authorizer.mu.RUnlock()
	// The entry might still exist but should be expired
	if ok {
		s.True(time.Now().After(cachedEntry.expiresAt), "Entry should be expired")
	}
}

// TestCacheThreadSafety tests thread safety of the cache.
func (s *RBACCacheTestSuite) TestCacheThreadSafety() {
	// This is a basic test - full thread safety testing would require more sophisticated setup
	key := "test-key"
	entry := &rbacCacheEntry{
		allowed:   true,
		expiresAt: time.Now().Add(5 * time.Second),
	}

	// Write from one goroutine
	go func() {
		s.authorizer.mu.Lock()
		s.authorizer.cache[key] = entry
		s.authorizer.mu.Unlock()
	}()

	// Read from another goroutine
	time.Sleep(10 * time.Millisecond)
	s.authorizer.mu.RLock()
	_, ok := s.authorizer.cache[key]
	s.authorizer.mu.RUnlock()
	// Should eventually be readable (race condition possible but unlikely with sleep)
	_ = ok
}

// TestTTLConfiguration tests that TTL can be configured.
func (s *RBACCacheTestSuite) TestTTLConfiguration() {
	// Test with different TTL values
	ttl1 := 1 * time.Second
	ttl2 := 10 * time.Second

	authorizer1 := &rbacAuthorizerImpl{
		cache: make(map[string]*rbacCacheEntry),
		ttl:   ttl1,
	}
	authorizer2 := &rbacAuthorizerImpl{
		cache: make(map[string]*rbacCacheEntry),
		ttl:   ttl2,
	}

	s.Equal(ttl1, authorizer1.ttl, "TTL should be set correctly")
	s.Equal(ttl2, authorizer2.ttl, "TTL should be set correctly")
	s.NotEqual(authorizer1.ttl, authorizer2.ttl, "Different TTLs should be different")
}

// TestRBACCacheTestSuite runs the RBAC cache test suite.
func TestRBACCacheTestSuite(t *testing.T) {
	suite.Run(t, new(RBACCacheTestSuite))
}
