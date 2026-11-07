package ctrls

import (
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/controllers/tenant/account/repositories"
	"booking-saas/controllers/tenant/account/usecases"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type RegisterController struct{ Usecase *usecases.RegisterUsecase }

func NewRegisterController(u *usecases.RegisterUsecase) *RegisterController {
	return &RegisterController{Usecase: u}
}
func (ctrl *RegisterController) Register(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "กรุณากรอกข้อมูลสมัครให้ครบถ้วนและถูกต้อง"})
		return
	}
	data, err := ctrl.Usecase.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrFreePackageUnavailable):
			c.JSON(503, gin.H{"success": false, "message": "แพ็กเกจเริ่มต้นยังไม่พร้อมใช้งาน กรุณาลองใหม่ภายหลัง"})
		case errors.Is(err, usecases.ErrInvalidRegistration):
			c.JSON(400, gin.H{"success": false, "message": "ข้อมูลสมัครไม่ถูกต้อง กรุณาตรวจสอบอีเมล เบอร์โทร และรหัสผ่าน"})
		case errors.Is(err, repositories.ErrAccountExists):
			c.JSON(409, gin.H{"success": false, "message": "ไม่สามารถสมัครด้วยข้อมูลบัญชีนี้ได้ กรุณาใช้อีเมลอื่น"})
		default:
			log.Printf("Tenant registration failed: %v", err)
			c.JSON(500, gin.H{"success": false, "message": "ไม่สามารถสมัครได้ในขณะนี้ กรุณาลองใหม่ภายหลัง"})
		}
		return
	}
	c.JSON(201, gin.H{"success": true, "message": "สมัครสมาชิกเรียบร้อยแล้ว สามารถเข้าสู่ระบบเพื่อเพิ่มร้านค้าได้", "data": data})
}
