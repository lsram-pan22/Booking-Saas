package routes

import (
	"net/http/httptest"
	"strings"
	"testing"

	"booking-saas/config"
	"github.com/gin-gonic/gin"
)

func TestTenantEndpointsRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter(&config.Config{JWT: config.JWTConfig{Secret: "route-test-secret"}})
	for _, tc := range []struct {
		method string
		path   string
		status int
	}{
		{"POST", "/api/v1/tenant/login", 400},
		{"POST", "/api/v1/tenant/register", 400},
		{"GET", "/api/v1/tenant/me", 401},
		{"PATCH", "/api/v1/tenant/me", 401},
		{"POST", "/api/v1/tenant/shops", 401},
		{"GET", "/api/v1/tenant/shops/11111111-1111-4111-8111-111111111111", 401},
		{"PATCH", "/api/v1/tenant/shops/11111111-1111-4111-8111-111111111111", 401},
		{"GET", "/api/v1/tenant/shops/11111111-1111-4111-8111-111111111111/members", 401},
		{"POST", "/api/v1/tenant/shops/11111111-1111-4111-8111-111111111111/members", 401},
		{"PATCH", "/api/v1/tenant/shops/11111111-1111-4111-8111-111111111111/members/22222222-2222-4222-8222-222222222222", 401},
	} {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d", response.Code, tc.status)
			}
		})
	}
}
