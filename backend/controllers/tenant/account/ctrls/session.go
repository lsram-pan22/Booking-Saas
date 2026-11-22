package ctrls

import (
	"booking-saas/middleware"
	"booking-saas/utils"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

// Version revocation intentionally logs out all devices for this account.
func LogoutAllSessions(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := middleware.AuthClaims(c)
		if !ok {
			c.AbortWithStatus(401)
			return
		}
		if db == nil {
			c.JSON(503, gin.H{"success": false, "message": "ฐานข้อมูลไม่พร้อมใช้งาน"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		revoked, err := utils.RevokeAccountSessions(ctx, db, claims)
		if err != nil {
			log.Printf("Session revocation failed: %v", err)
			c.JSON(503, gin.H{"success": false, "message": "ออกจากระบบทุกอุปกรณ์ไม่สำเร็จ กรุณาลองใหม่"})
			return
		}
		if !revoked {
			c.JSON(401, gin.H{"success": false, "message": "Session ถูกยกเลิกแล้ว"})
			return
		}
		c.JSON(200, gin.H{"success": true})
	}
}
