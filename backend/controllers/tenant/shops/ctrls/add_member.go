package ctrls

import (
	"booking-saas/controllers/tenant/shops/models"
	"booking-saas/controllers/tenant/shops/repositories"
	"booking-saas/controllers/tenant/shops/usecases"
	"booking-saas/middleware"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type AddMemberController struct{ Usecase *usecases.AddMemberUsecase }

func NewAddMemberController(u *usecases.AddMemberUsecase) *AddMemberController {
	return &AddMemberController{Usecase: u}
}
func (ctrl *AddMemberController) Add(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.AddMemberRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"success": false, "message": "กรุณากรอกข้อมูลผู้ใช้งานให้ครบถ้วน"})
		return
	}
	member, err := ctrl.Usecase.Add(c.Request.Context(), claims.UserID, c.Param("shopID"), req)
	if err != nil {
		status, message := 500, "เพิ่มผู้ใช้งานไม่สำเร็จ กรุณาลองใหม่"
		switch {
		case errors.Is(err, usecases.ErrInvalidMember):
			status, message = 400, "กรุณาตรวจสอบข้อมูลผู้ใช้งานและรหัสผ่าน"
		case errors.Is(err, repositories.ErrShopNotFound):
			status, message = 404, "ไม่พบร้านค้าที่เปิดใช้งานหรือคุณไม่มีสิทธิ์เพิ่มผู้ใช้งาน"
		case errors.Is(err, repositories.ErrSubscriptionRequired):
			status, message = 403, "ต้องมีแพ็กเกจที่เปิดใช้งานก่อนเพิ่มผู้ใช้งาน"
		case errors.Is(err, repositories.ErrMemberQuota):
			status, message = 409, "จำนวนผู้ใช้งานถึงโควตาแพ็กเกจแล้ว (รวมเจ้าของร้าน)"
		case errors.Is(err, repositories.ErrMemberExists):
			status, message = 409, "บัญชีนี้เป็นสมาชิกของร้านอยู่แล้ว"
		case errors.Is(err, repositories.ErrStaffShopAssigned):
			status, message = 409, "พนักงานบัญชีนี้ผูกกับร้านอื่นแล้ว พนักงานหนึ่งคนอยู่ได้เพียงร้านเดียว"
		case errors.Is(err, repositories.ErrMemberEmailExists):
			status, message = 409, "อีเมลนี้มีบัญชีแล้ว กรุณาเลือกเพิ่มบัญชีที่มีอยู่"
		case errors.Is(err, repositories.ErrMemberUsernameExists):
			status, message = 409, "Username นี้ถูกใช้งานแล้ว กรุณาเลือกชื่ออื่น"
		case errors.Is(err, repositories.ErrMemberAccountUnavailable):
			status, message = 400, "ไม่พบบัญชีที่พร้อมเพิ่มเข้าร้านด้วยอีเมลนี้"
		default:
			log.Printf("Add shop member: %v", err)
		}
		c.JSON(status, gin.H{"success": false, "message": message})
		return
	}
	c.JSON(201, gin.H{"success": true, "data": member})
}
