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

type MemberStatusController struct{ Usecase *usecases.MemberStatusUsecase }

func NewMemberStatusController(u *usecases.MemberStatusUsecase) *MemberStatusController {
	return &MemberStatusController{Usecase: u}
}
func (ctrl *MemberStatusController) Set(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)
	var req models.MemberStatusRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"success": false, "message": "สถานะสมาชิกไม่ถูกต้อง"})
		return
	}
	data, err := ctrl.Usecase.Set(c.Request.Context(), claims.UserID, c.Param("shopID"), c.Param("memberID"), req)
	if err != nil {
		status, message := 500, "เปลี่ยนสถานะสมาชิกไม่สำเร็จ"
		switch {
		case errors.Is(err, usecases.ErrInvalidMember):
			status, message = 400, "ข้อมูลหรือสถานะสมาชิกไม่ถูกต้อง"
		case errors.Is(err, repositories.ErrShopNotFound):
			status, message = 404, "ไม่พบสมาชิกในร้านหรือคุณไม่มีสิทธิ์จัดการ"
		case errors.Is(err, repositories.ErrMemberProtected):
			status, message = 403, "ไม่สามารถเปลี่ยนสถานะเจ้าของร้านหรือสมาชิกที่ถูกระงับได้"
		case errors.Is(err, repositories.ErrSubscriptionRequired):
			status, message = 403, "ต้องมีแพ็กเกจที่เปิดใช้งานก่อนเปิดสมาชิก"
		case errors.Is(err, repositories.ErrMemberQuota):
			status, message = 409, "จำนวนผู้ใช้งานถึงโควตาแพ็กเกจแล้ว"
		case errors.Is(err, repositories.ErrMemberAccountUnavailable):
			status, message = 403, "บัญชีผู้ใช้งานนี้ถูกปิดหรือระงับ จึงเปิดสมาชิกไม่ได้"
		default:
			log.Printf("Set shop member status: %v", err)
		}
		c.JSON(status, gin.H{"success": false, "message": message})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
