package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimiterConcurrencyExpiryAndCapacity(t *testing.T) {
	limiter := NewRateLimiter(7, time.Minute)
	var wg sync.WaitGroup
	var accepted atomic.Int32
	for range 40 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _ := limiter.allow("same"); ok {
				accepted.Add(1)
			}
		}()
	}
	wg.Wait()
	if accepted.Load() != 7 {
		t.Fatalf("accepted %d", accepted.Load())
	}
	now := time.Now()
	limiter = NewRateLimiter(1, time.Minute)
	limiter.now = func() time.Time { return now }
	limiter.maxKeys = 2
	if ok, _ := limiter.allow("a"); !ok {
		t.Fatal("first request denied")
	}
	limiter.allow("b")
	if ok, _ := limiter.allow("c"); ok || len(limiter.entries) > 2 {
		t.Fatal("unbounded limiter")
	}
	now = now.Add(2 * time.Minute)
	if ok, _ := limiter.allow("c"); !ok {
		t.Fatal("expired counters did not clear")
	}
}
func TestIdentityLimitAcrossIPsAndLoginSurfaces(t *testing.T) {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(AuthenticationLimits())
	for _, path := range []string{"/api/v1/tenant/login", "/api/v1/super-admin/login"} {
		router.POST(path, func(c *gin.Context) {
			var req struct {
				Username string `json:"username"`
			}
			if DecodeJSON(c, &req) != nil {
				c.Status(400)
				return
			}
			c.Status(401)
		})
	}
	for i := 0; i < 11; i++ {
		body := `{"username":" Owner@Example.test "}`
		if i%2 == 0 {
			body = `{"username":"owner@example.test"}`
		}
		path := "/api/v1/tenant/login"
		if i%2 == 0 {
			path = "/api/v1/super-admin/login"
		}
		req := httptest.NewRequest("POST", path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = fmt.Sprintf("10.0.0.%d:1234", i+1)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		expected := 401
		if i == 10 {
			expected = 429
			if rec.Header().Get("Retry-After") == "" {
				t.Fatal("missing retry information")
			}
		}
		if rec.Code != expected {
			t.Fatalf("request %d: %d", i, rec.Code)
		}
	}
}
func TestSpoofedForwardingHeaderCannotBypassIPLimit(t *testing.T) {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(AuthenticationLimits())
	router.POST("/api/v1/tenant/login", func(c *gin.Context) { c.Status(400) })
	for i := 0; i < 31; i++ {
		req := httptest.NewRequest("POST", "/api/v1/tenant/login", strings.NewReader(fmt.Sprintf(`{"username":"user-%d"}`, i)))
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("192.0.2.%d", i+1))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		expected := 400
		if i == 30 {
			expected = 429
		}
		if rec.Code != expected {
			t.Fatalf("IP limit bypass: %d", rec.Code)
		}
	}
}
