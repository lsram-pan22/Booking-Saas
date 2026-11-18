package middleware

import (
	"bytes"
	"crypto/tls"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStrictJSONRejectsPrivilegeFieldsAndAmbiguity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body, contentType string
		status                  int
	}{
		{"valid", `{"name":"Test"}`, "application/json", 200},
		{"unknown role", `{"name":"Test","role":"owner"}`, "application/json", 400},
		{"two documents", `{"name":"Test"}{"name":"Other"}`, "application/json", 400},
		{"missing required", `{}`, "application/json", 400},
		{"null", `null`, "application/json", 400},
		{"array", `[]`, "application/json", 400},
		{"wrong media", `{"name":"Test"}`, "text/plain", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(LimitRequestBody())
			router.POST("/", func(c *gin.Context) {
				var input struct {
					Name string `json:"name" binding:"required"`
				}
				if DecodeJSON(c, &input) != nil {
					c.Status(400)
					return
				}
				c.Status(200)
			})
			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Fatalf("status %d want %d", rec.Code, tc.status)
			}
		})
	}
}
func TestSecurityHeadersAndBodyBounds(t *testing.T) {
	router := gin.New()
	router.Use(SecurityHeaders(), LimitRequestBody())
	router.POST("/", func(c *gin.Context) {
		if _, err := io.ReadAll(c.Request.Body); err != nil {
			c.Status(413)
			return
		}
		c.Status(200)
	})
	for _, chunked := range []bool{false, true} {
		req := httptest.NewRequest("POST", "/", strings.NewReader(strings.Repeat("x", int(MaxRequestBody)+1)))
		if chunked {
			req.ContentLength = -1
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != 413 {
			t.Fatalf("oversize status %d", rec.Code)
		}
		if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatal("security headers missing")
		}
		if rec.Header().Get("Strict-Transport-Security") != "" {
			t.Fatal("HSTS set on HTTP")
		}
	}
	req := httptest.NewRequest("POST", "/", nil)
	req.TLS = &tls.ConnectionState{}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Header().Get("Strict-Transport-Security") == "" {
		t.Fatal("missing HTTPS HSTS")
	}
}
func TestSecurityAuditDoesNotLogCredentialsOrQuery(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	router := gin.New()
	router.Use(SecurityAudit())
	router.POST("/login", func(c *gin.Context) { c.Status(401) })
	req := httptest.NewRequest("POST", "/login?password=private-query", strings.NewReader(`{"password":"private-body"}`))
	req.Header.Set("Authorization", "Bearer private-token")
	req.Header.Set("X-Request-ID", "private-client-id")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	for _, secret := range []string{"private-query", "private-body", "private-token", "private-client-id"} {
		if strings.Contains(output.String(), secret) {
			t.Fatal("secret in security log")
		}
	}
	if !strings.Contains(output.String(), `"route":"/login"`) || len(rec.Header().Get("X-Request-ID")) != 32 {
		t.Fatal("missing audit correlation")
	}
}

func TestPanicDoesNotDumpSecrets(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	router := gin.New()
	router.Use(SecurityAudit(), SafeRecovery())
	router.GET("/panic", func(*gin.Context) { panic("private-panic-value") })
	req := httptest.NewRequest("GET", "/panic?secret=private-query", nil)
	req.Header.Set("Authorization", "Bearer private-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatal("panic was not contained")
	}
	for _, secret := range []string{"private-query", "private-token", "private-panic-value"} {
		if strings.Contains(output.String(), secret) || strings.Contains(rec.Body.String(), secret) {
			t.Fatal("panic leaked request information")
		}
	}
}
