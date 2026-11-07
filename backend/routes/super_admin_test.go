package routes

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"booking-saas/config"
	"booking-saas/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func expiredAdminToken(t *testing.T, secret string) string {
	t.Helper()
	claims := utils.JWTClaims{
		UserID: 1, Role: utils.RoleSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "booking-saas",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestPackageRoutesRejectNonSuperAdmins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "package-route-test-secret"
	cfg := &config.Config{JWT: config.JWTConfig{Secret: secret}}
	router := gin.New()
	// A nil database ensures denied requests cannot reach repository operations.
	SuperAdminRoutes(router.Group("/api"), nil, cfg)
	endpoints := []struct{ method, path string }{
		{"GET", "/api/super-admin/packages"},
		{"POST", "/api/super-admin/packages"},
		{"PUT", "/api/super-admin/packages/2cd5d4ae-871b-46f6-ab6e-b2d87d51c231"},
	}
	tests := []struct {
		name, role string
		expires    time.Duration
		status     int
	}{
		{"no token", "", 0, 401},
		{"owner", utils.RoleOwner, time.Hour, 403},
		{"staff", utils.RoleStaff, time.Hour, 403},
		{"unknown role", "admin", time.Hour, 403},
		{"empty role", "", time.Hour, 403},
		{"expired super admin", utils.RoleSuperAdmin, -time.Hour, 401},
	}
	for _, endpoint := range endpoints {
		for _, tt := range tests {
			t.Run(endpoint.method+"/"+tt.name, func(t *testing.T) {
				req := httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader(`{}`))
				req.Header.Set("Content-Type", "application/json")
				if tt.expires != 0 {
					token, err := utils.GenerateJWT(secret, time.Hour, 1, "user@example.com", tt.role)
					if err != nil {
						t.Fatal(err)
					}
					if tt.expires < 0 {
						token = expiredAdminToken(t, secret)
					}
					req.Header.Set("Authorization", "Bearer "+token)
				}
				response := httptest.NewRecorder()
				router.ServeHTTP(response, req)
				if response.Code != tt.status {
					t.Fatalf("status = %d, want %d", response.Code, tt.status)
				}
			})
		}
	}

	// A valid admin reaches request validation (400), rather than authorization denial.
	token, err := utils.GenerateJWT(secret, time.Hour, 1, "admin@example.com", utils.RoleSuperAdmin)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/super-admin/packages", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != 400 {
		t.Fatalf("valid admin status = %d, want validation status 400", response.Code)
	}
}
