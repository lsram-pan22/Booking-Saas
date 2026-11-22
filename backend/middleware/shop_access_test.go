package middleware

import (
	"booking-saas/utils"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShopBoundaryAndOwnerOnlyOperations(t *testing.T) {
	const ownShop = "11111111-1111-4111-8111-111111111111"
	const otherShop = "22222222-2222-4222-8222-222222222222"
	for _, tc := range []struct {
		name, actor, shop, role string
		ownerOnly, dbFailure    bool
		want                    int
	}{
		{"owner allowed", "owner", ownShop, "user", true, false, 200},
		{"staff booking allowed", "staff", ownShop, "user", false, false, 200},
		{"staff management denied", "staff", ownShop, "user", true, false, 404},
		{"cross tenant denied", "owner", otherShop, "user", false, false, 404},
		{"invalid shop denied", "owner", "not-a-uuid", "user", false, false, 404},
		{"super admin not tenant member", "owner", ownShop, "super_admin", false, false, 403},
		{"database failure closed", "owner", ownShop, "user", false, true, 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set(ClaimsKey, &utils.JWTClaims{UserID: tc.actor, Role: tc.role}) })
			lookup := func(_ context.Context, actor, shop string, ownerOnly bool) (bool, error) {
				if tc.dbFailure {
					return false, errors.New("unavailable")
				}
				return shop == ownShop && (actor == "owner" || (!ownerOnly && actor == "staff")), nil
			}
			router.POST("/shops/:shopID", shopAccess(lookup, tc.ownerOnly), func(c *gin.Context) { called = true; c.Status(200) })
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/shops/"+tc.shop+"?user_id=owner", strings.NewReader(`{"user_id":"owner","role":"owner"}`))
			router.ServeHTTP(rec, req)
			if rec.Code != tc.want || called != (tc.want == 200) {
				t.Fatalf("boundary status %d, called %v", rec.Code, called)
			}
		})
	}
}
