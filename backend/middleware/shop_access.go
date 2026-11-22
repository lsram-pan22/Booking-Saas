package middleware

import (
	"booking-saas/utils"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type shopLookup func(context.Context, string, string, bool) (bool, error)

func shopAccess(lookup shopLookup, ownerOnly bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := AuthClaims(c)
		if !ok {
			unauthorized(c)
			return
		}
		if claims.Role != "user" {
			c.AbortWithStatusJSON(403, gin.H{"success": false, "message": "ไม่มีสิทธิ์เข้าใช้งานร้านค้า"})
			return
		}
		shopID := c.Param("shopID")
		if !utils.ValidUserID(shopID) {
			c.AbortWithStatusJSON(404, gin.H{"success": false, "message": "ไม่พบร้านค้าหรือไม่มีสิทธิ์เข้าถึง"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		allowed, err := lookup(ctx, claims.UserID, shopID, ownerOnly)
		if err != nil {
			c.AbortWithStatusJSON(503, gin.H{"success": false, "message": "ตรวจสอบสิทธิ์ร้านค้าไม่สำเร็จ"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(404, gin.H{"success": false, "message": "ไม่พบร้านค้าหรือไม่มีสิทธิ์เข้าถึง"})
			return
		}
		c.Next()
	}
}

// A second boundary in addition to transaction-local repository checks. Every
// new shop route must opt into this guard; assignment checks stay in SQL.
func RequireShopAccess(db *pgxpool.Pool, ownerOnly bool) gin.HandlerFunc {
	return shopAccess(func(ctx context.Context, userID, shopID string, ownerOnly bool) (bool, error) {
		if db == nil {
			return false, context.Canceled
		}
		var allowed bool
		err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_members m JOIN tenants t ON t.id=m.tenant_id JOIN users u ON u.id=m.user_id
 WHERE t.id=$1 AND m.user_id=$2 AND m.status='active' AND u.status='active' AND u.system_role='user' AND m.role IN ('owner','staff') AND (NOT $3 OR m.role='owner') AND (t.status='active' OR (m.role='owner' AND t.status IN ('inactive','pending'))))`, shopID, userID, ownerOnly).Scan(&allowed)
		return allowed, err
	}, ownerOnly)
}
