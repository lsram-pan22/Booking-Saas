package ctrls

import (
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/controllers/tenant/account/usecases"
	"booking-saas/middleware"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"time"
)

type PasswordController struct{ Usecase *usecases.PasswordUsecase }

func NewPasswordController(u *usecases.PasswordUsecase) *PasswordController {
	return &PasswordController{Usecase: u}
}
func passwordActor(c *gin.Context) models.PasswordActor {
	claims, _ := middleware.AuthClaims(c)
	return models.PasswordActor{ID: claims.UserID, Role: claims.Role, Version: claims.SessionVersion}
}
func passwordResult(c *gin.Context, err error) {
	switch {
	case err == nil:
		c.JSON(200, gin.H{"success": true})
	case errors.Is(err, usecases.ErrInvalidPassword):
		c.JSON(400, gin.H{"success": false, "message": "ใช้รหัสผ่านอย่างน้อย 12 ตัวอักษร และไม่เกิน 72 ไบต์"})
	case errors.Is(err, usecases.ErrCurrentPassword):
		c.JSON(400, gin.H{"success": false, "message": "รหัสผ่านปัจจุบันไม่ถูกต้อง หรือรหัสใหม่ซ้ำกับรหัสเดิม"})
	case errors.Is(err, usecases.ErrPasswordDenied):
		c.JSON(403, gin.H{"success": false, "message": "ไม่สามารถเปลี่ยนรหัสบัญชีนี้ได้ กรุณาตรวจสอบสิทธิ์และสถานะบัญชี"})
	default:
		log.Printf("Password operation failed: %v", err)
		c.JSON(500, gin.H{"success": false, "message": "เปลี่ยนรหัสผ่านไม่สำเร็จ"})
	}
}
func (ctrl *PasswordController) Reset(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
	var req models.ResetPasswordRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	passwordResult(c, ctrl.Usecase.Reset(c.Request.Context(), passwordActor(c), c.Param("shopID"), c.Param("userID"), req))
}
func (ctrl *PasswordController) Change(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
	var req models.ChangePasswordRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	passwordResult(c, ctrl.Usecase.Change(c.Request.Context(), passwordActor(c), req))
}
func (ctrl *PasswordController) Owners(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 254 {
		c.JSON(400, gin.H{"success": false, "message": "คำค้นหายาวเกินไป"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	items, err := ctrl.Usecase.Repo.Owners(ctx, search)
	if err != nil {
		passwordResult(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": items})
}
