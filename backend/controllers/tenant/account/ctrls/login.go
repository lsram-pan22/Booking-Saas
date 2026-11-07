package ctrls

import (
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/controllers/tenant/account/usecases"
	"booking-saas/middleware"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type LoginController struct{ Usecase *usecases.LoginUsecase }

func NewLoginController(u *usecases.LoginUsecase) *LoginController {
	return &LoginController{Usecase: u}
}
func (ctrl *LoginController) Login(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.LoginRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"success": false, "message": "กรุณากรอกอีเมลและรหัสผ่านให้ถูกต้อง"})
		return
	}
	data, err := ctrl.Usecase.Login(c.Request.Context(), req)
	if err != nil {
		loginError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
func (ctrl *LoginController) Me(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	data, err := ctrl.Usecase.Profile(c.Request.Context(), claims.UserID)
	if err != nil {
		loginError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
func loginError(c *gin.Context, err error) {
	if errors.Is(err, usecases.ErrInvalidCredentials) {
		c.JSON(401, gin.H{"success": false, "message": "อีเมลหรือรหัสผ่านไม่ถูกต้อง หรือบัญชีถูกระงับ"})
		return
	}
	log.Printf("Tenant authentication failed: %v", err)
	c.JSON(500, gin.H{"success": false, "message": "ไม่สามารถเข้าสู่ระบบได้ในขณะนี้"})
}
