package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"booking-saas/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestSuperAdminAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "test-secret-for-jwt-authorization"
	now := time.Now()
	base := utils.JWTClaims{
		UserID: "11111111-1111-4111-8111-111111111111", Role: "super_admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "booking-saas", IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
	sign := func(claims utils.JWTClaims, method jwt.SigningMethod, key string) string {
		t.Helper()
		token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	valid := sign(base, jwt.SigningMethodHS256, secret)
	tests := []struct {
		name   string
		header string
		mutate func(*utils.JWTClaims)
		method jwt.SigningMethod
		key    string
		status int
	}{
		{name: "missing token", status: 401},
		{name: "malformed token", header: "Bearer invalid", status: 401},
		{name: "wrong scheme", header: "Basic " + valid, status: 401},
		{name: "extra fields", header: "Bearer " + valid + " extra", status: 401},
		{name: "valid admin", header: "Bearer " + valid, status: 200},
		{name: "case insensitive scheme", header: "bearer " + valid, status: 200},
		{name: "wrong signature", key: "other-secret", status: 401},
		{name: "wrong algorithm", method: jwt.SigningMethodHS384, status: 401},
		{name: "expired", mutate: func(c *utils.JWTClaims) { c.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Hour)) }, status: 401},
		{name: "missing expiry", mutate: func(c *utils.JWTClaims) { c.ExpiresAt = nil }, status: 401},
		{name: "wrong issuer", mutate: func(c *utils.JWTClaims) { c.Issuer = "other" }, status: 401},
		{name: "future issued", mutate: func(c *utils.JWTClaims) { c.IssuedAt = jwt.NewNumericDate(now.Add(time.Hour)) }, status: 401},
		{name: "invalid user", mutate: func(c *utils.JWTClaims) { c.UserID = "invalid" }, status: 401},
		{name: "wrong role", mutate: func(c *utils.JWTClaims) { c.Role = "admin" }, status: 403},
		{name: "missing role", mutate: func(c *utils.JWTClaims) { c.Role = "" }, status: 403},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := tt.header
			if tt.mutate != nil || tt.method != nil || tt.key != "" {
				claims := base
				if tt.mutate != nil {
					tt.mutate(&claims)
				}
				method := tt.method
				if method == nil {
					method = jwt.SigningMethodHS256
				}
				key := tt.key
				if key == "" {
					key = secret
				}
				header = "Bearer " + sign(claims, method, key)
			}
			router := gin.New()
			called := false
			router.GET("/protected", RequireJWT(secret), RequireRole("super_admin"), func(c *gin.Context) {
				called = true
				claims := c.MustGet(ClaimsKey).(*utils.JWTClaims)
				if claims.UserID != "11111111-1111-4111-8111-111111111111" {
					t.Error("unexpected authenticated user")
				}
				c.Status(200)
			})
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", header)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d", response.Code, tt.status)
			}
			if called != (tt.status == 200) {
				t.Fatal("protected handler execution mismatch")
			}
		})
	}
}

func TestMissingAuthenticationFailsClosed(t *testing.T) {
	for _, handler := range []gin.HandlerFunc{RequireJWT(""), RequireRole("super_admin")} {
		router := gin.New()
		router.GET("/", handler, func(c *gin.Context) { t.Error("unauthenticated handler ran") })
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
		if response.Code != 401 {
			t.Fatalf("status = %d, want 401", response.Code)
		}
	}
}

func TestSharedRoleAuthorization(t *testing.T) {
	const secret = "shared-role-test-secret"
	roles := []string{utils.RoleSuperAdmin, utils.RoleOwner, utils.RoleStaff}
	policies := []struct {
		name    string
		allowed []string
	}{
		{"super-admin", []string{utils.RoleSuperAdmin}},
		{"owner", []string{utils.RoleOwner}},
		{"staff", []string{utils.RoleStaff}},
		{"team", []string{utils.RoleOwner, utils.RoleStaff}},
		{"no-roles", nil},
		{"empty-role", []string{""}},
	}
	for _, policy := range policies {
		for _, role := range roles {
			t.Run(policy.name+"/"+role, func(t *testing.T) {
				token, err := utils.GenerateJWT(secret, time.Hour, "42424242-4242-4242-8242-424242424242", "user@example.com", role)
				if err != nil {
					t.Fatal(err)
				}
				want := 403
				for _, allowed := range policy.allowed {
					if role == allowed {
						want = 200
					}
				}
				router := gin.New()
				router.GET("/", RequireJWT(secret), RequireRole(policy.allowed...), func(c *gin.Context) {
					claims, ok := AuthClaims(c)
					if !ok || claims.UserID != "42424242-4242-4242-8242-424242424242" || claims.Role != role {
						t.Error("incorrect verified claims")
					}
					c.Status(200)
				})
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, req)
				if response.Code != want {
					t.Fatalf("status = %d, want %d", response.Code, want)
				}
			})
		}
	}
}
