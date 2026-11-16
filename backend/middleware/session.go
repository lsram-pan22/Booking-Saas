package middleware

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

type SessionState struct {
	Version    int64
	Role       string
	Status     string
	MustChange bool
}

// SessionPolicy is kept separate from storage so revocation and restricted
// temporary-password sessions can be verified without a live database.
func SessionPolicy(state SessionState, version int64, role string, allowChange bool) int {
	if state.Status != "active" || state.Role != role || state.Version != version {
		return 401
	}
	if state.MustChange && !allowChange {
		return 403
	}
	return 0
}

func RequireSession(db *pgxpool.Pool, allowChange bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := AuthClaims(c)
		if !ok {
			unauthorized(c)
			return
		}
		if db == nil {
			c.AbortWithStatusJSON(503, gin.H{"success": false, "message": "ฐานข้อมูลไม่พร้อมใช้งาน"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		var state SessionState
		err := db.QueryRow(ctx, `SELECT session_version,system_role,status,must_change_password FROM users WHERE id=$1`, claims.UserID).Scan(&state.Version, &state.Role, &state.Status, &state.MustChange)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				unauthorized(c)
			} else {
				log.Printf("Session lookup failed: %v", err)
				c.AbortWithStatusJSON(503, gin.H{"success": false, "message": "ตรวจสอบ session ไม่สำเร็จ กรุณาลองใหม่"})
			}
			return
		}
		switch SessionPolicy(state, claims.SessionVersion, claims.Role, allowChange) {
		case 401:
			unauthorized(c)
			return
		case 403:
			c.AbortWithStatusJSON(403, gin.H{"success": false, "code": "PASSWORD_CHANGE_REQUIRED", "message": "กรุณาเปลี่ยนรหัสผ่านชั่วคราวก่อนใช้งาน"})
			return
		}
		c.Next()
	}
}
