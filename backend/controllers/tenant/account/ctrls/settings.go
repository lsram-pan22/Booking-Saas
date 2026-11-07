package ctrls

import (
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/controllers/tenant/account/usecases"
	"booking-saas/middleware"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
)

type SettingsController struct{ Usecase *usecases.SettingsUsecase }

func NewSettingsController(u *usecases.SettingsUsecase) *SettingsController {
	return &SettingsController{Usecase: u}
}
func (ctrl *SettingsController) Update(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.UpdateProfileRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	user, err := ctrl.Usecase.Update(c.Request.Context(), claims.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrInvalidRegistration):
			c.JSON(400, gin.H{"success": false, "message": "กรุณาตรวจสอบชื่อและเบอร์โทรศัพท์"})
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(401, gin.H{"success": false, "message": "บัญชีไม่พร้อมใช้งาน กรุณาเข้าสู่ระบบใหม่"})
		default:
			log.Printf("Update tenant profile: %v", err)
			c.JSON(500, gin.H{"success": false, "message": "บันทึกข้อมูลไม่สำเร็จ"})
		}
		return
	}
	c.JSON(200, gin.H{"success": true, "data": user})
}
