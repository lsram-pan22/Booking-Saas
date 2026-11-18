package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateEntry struct {
	count int
	until time.Time
}

func AuthenticationLimits() gin.HandlerFunc {
	ip := NewRateLimiter(30, 5*time.Minute)
	identity := NewRateLimiter(10, 5*time.Minute)
	registration := NewRateLimiter(5, time.Hour)
	return func(c *gin.Context) {
		if c.Request.Method == "POST" {
			switch c.FullPath() {
			case "/api/v1/tenant/login", "/api/v1/super-admin/login":
				if ok, retry := ip.allow("ip:" + c.ClientIP()); !ok {
					denyRate(c, retry)
					return
				}
				identity.ByLoginIdentity()(c)
				return
			case "/api/v1/tenant/register":
				if ok, retry := registration.allow("ip:" + c.ClientIP()); !ok {
					denyRate(c, retry)
					return
				}
			}
		}
		c.Next()
	}
}

type RateLimiter struct {
	mu             sync.Mutex
	entries        map[string]rateEntry
	limit, maxKeys int
	window         time.Duration
	now            func() time.Time
	nextCleanup    time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{entries: map[string]rateEntry{}, limit: limit, maxKeys: 10000, window: window, now: time.Now}
}
func (l *RateLimiter) allow(key string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if !now.Before(l.nextCleanup) {
		for k, v := range l.entries {
			if !now.Before(v.until) {
				delete(l.entries, k)
			}
		}
		l.nextCleanup = now.Add(time.Minute)
	}
	entry, exists := l.entries[key]
	if !exists || !now.Before(entry.until) {
		if !exists && len(l.entries) >= l.maxKeys {
			return false, 60
		}
		entry = rateEntry{until: now.Add(l.window)}
	}
	if entry.count >= l.limit {
		return false, int(math.Ceil(entry.until.Sub(now).Seconds()))
	}
	entry.count++
	l.entries[key] = entry
	return true, 0
}
func denyRate(c *gin.Context, seconds int) {
	if seconds < 1 {
		seconds = 1
	}
	c.Header("Retry-After", strconv.Itoa(seconds))
	c.AbortWithStatusJSON(429, gin.H{"success": false, "message": "ทำรายการถี่เกินไป กรุณารอสักครู่แล้วลองใหม่"})
}
func (l *RateLimiter) ByIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ok, retry := l.allow("ip:" + c.ClientIP()); !ok {
			denyRate(c, retry)
			return
		}
		c.Next()
	}
}
func (l *RateLimiter) ByActor() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := AuthClaims(c)
		if !ok {
			unauthorized(c)
			return
		}
		if allowed, retry := l.allow("actor:" + claims.UserID); !allowed {
			denyRate(c, retry)
			return
		}
		c.Next()
	}
}

// Pair with ByIP before this middleware. The normalized identity is hashed in
// memory; use the same limiter across tenant/admin login entry points.
func (l *RateLimiter) ByLoginIdentity() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, 16385))
		if err != nil || len(body) > 16384 {
			c.AbortWithStatusJSON(413, gin.H{"success": false, "message": "ข้อมูลเข้าสู่ระบบมีขนาดใหญ่เกินไป"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		var identity struct {
			Username string `json:"username"`
			Email    string `json:"email"`
		}
		if json.Unmarshal(body, &identity) == nil {
			value := strings.ToLower(strings.TrimSpace(identity.Username))
			if value == "" {
				value = strings.ToLower(strings.TrimSpace(identity.Email))
			}
			if value != "" {
				digest := sha256.Sum256([]byte(value))
				if ok, retry := l.allow("identity:" + hex.EncodeToString(digest[:])); !ok {
					denyRate(c, retry)
					return
				}
			}
		}
		c.Next()
	}
}
