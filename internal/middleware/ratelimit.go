package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// visitorEntry holds a limiter and its last access time
type visitorEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiter implements a rate limiter using token bucket algorithm
type RateLimiter struct {
	visitors map[string]*visitorEntry
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	ttl      time.Duration // How long to keep idle entries
	stopCh   chan struct{}
}

// NewRateLimiter creates a new rate limiter
// r: requests per second
// burst: maximum burst size
func NewRateLimiter(r float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitorEntry),
		rate:     rate.Limit(r),
		burst:    burst,
		ttl:      5 * time.Minute, // Default TTL
		stopCh:   make(chan struct{}),
	}
	// Start automatic cleanup
	go rl.startAutoCleanup()
	return rl
}

// NewRateLimiterWithTTL creates a rate limiter with custom TTL
func NewRateLimiterWithTTL(r float64, burst int, ttl time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitorEntry),
		rate:     rate.Limit(r),
		burst:    burst,
		ttl:      ttl,
		stopCh:   make(chan struct{}),
	}
	go rl.startAutoCleanup()
	return rl
}

// GetLimiter returns a limiter for a specific identifier (user ID, API key, IP)
func (rl *RateLimiter) GetLimiter(identifier string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.visitors[identifier]
	if !exists {
		entry = &visitorEntry{
			limiter:    rate.NewLimiter(rl.rate, rl.burst),
			lastAccess: time.Now(),
		}
		rl.visitors[identifier] = entry
	} else {
		entry.lastAccess = time.Now()
	}

	return entry.limiter
}

// startAutoCleanup periodically removes stale entries
func (rl *RateLimiter) startAutoCleanup() {
	ticker := time.NewTicker(rl.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupStale()
		case <-rl.stopCh:
			return
		}
	}
}

// cleanupStale removes entries that haven't been accessed within TTL
func (rl *RateLimiter) cleanupStale() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-rl.ttl)
	for key, entry := range rl.visitors {
		if entry.lastAccess.Before(threshold) {
			delete(rl.visitors, key)
		}
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// Count returns the current number of tracked visitors
func (rl *RateLimiter) Count() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.visitors)
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get user ID from context (if authenticated)
		identifier, exists := c.Get(AuthUserKey)
		if !exists {
			// Fall back to IP address for unauthenticated requests
			identifier = c.ClientIP()
		}

		identifierStr, ok := identifier.(string)
		if !ok {
			identifierStr = c.ClientIP()
		}

		rateLimiter := limiter.GetLimiter(identifierStr)
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PerUserRateLimiter implements different rate limits per role
type PerUserRateLimiter struct {
	adminLimiter    *RateLimiter
	userLimiter     *RateLimiter
	readonlyLimiter *RateLimiter
	defaultLimiter  *RateLimiter
}

// NewPerUserRateLimiter creates a rate limiter with different limits per role
func NewPerUserRateLimiter(adminRate, userRate, readonlyRate, defaultRate float64, burst int) *PerUserRateLimiter {
	return &PerUserRateLimiter{
		adminLimiter:    NewRateLimiter(adminRate, burst),
		userLimiter:     NewRateLimiter(userRate, burst),
		readonlyLimiter: NewRateLimiter(readonlyRate, burst),
		defaultLimiter:  NewRateLimiter(defaultRate, burst),
	}
}

// Stop stops all cleanup goroutines
func (rl *PerUserRateLimiter) Stop() {
	rl.adminLimiter.Stop()
	rl.userLimiter.Stop()
	rl.readonlyLimiter.Stop()
	rl.defaultLimiter.Stop()
}

// PerUserRateLimitMiddleware creates a role-based rate limiting middleware
func PerUserRateLimitMiddleware(limiter *PerUserRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier, exists := c.Get(AuthUserKey)
		if !exists {
			identifier = c.ClientIP()
		}

		identifierStr, ok := identifier.(string)
		if !ok {
			identifierStr = c.ClientIP()
		}

		// Select limiter based on role
		var rateLimiter *rate.Limiter
		role, roleExists := c.Get(AuthRoleKey)
		if roleExists {
			switch role {
			case "admin":
				rateLimiter = limiter.adminLimiter.GetLimiter(identifierStr)
			case "user":
				rateLimiter = limiter.userLimiter.GetLimiter(identifierStr)
			case "readonly":
				rateLimiter = limiter.readonlyLimiter.GetLimiter(identifierStr)
			default:
				rateLimiter = limiter.defaultLimiter.GetLimiter(identifierStr)
			}
		} else {
			rateLimiter = limiter.defaultLimiter.GetLimiter(identifierStr)
		}

		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
