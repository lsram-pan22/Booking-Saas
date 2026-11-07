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

type ShopController struct{ Usecase *usecases.ShopUsecase }

func NewShopController(u *usecases.ShopUsecase) *ShopController { return &ShopController{Usecase: u} }
func (ctrl *ShopController) Create(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.CreateShopRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"success": false, "message": "กรุณากรอกชื่อและประเภทร้านค้า"})
		return
	}
	shop, err := ctrl.Usecase.Create(c.Request.Context(), claims.UserID, req)
	if err != nil {
		status, message := 500, "เพิ่มร้านค้าไม่สำเร็จ กรุณาลองใหม่"
		switch {
		case errors.Is(err, repositories.ErrOwnerAccountRequired):
			status, message = 403, "บัญชีพนักงานไม่มีสิทธิ์สร้างร้านค้า"
		case errors.Is(err, usecases.ErrInvalidShop):
			status, message = 400, "ข้อมูลร้านค้าไม่ถูกต้อง"
		case errors.Is(err, repositories.ErrAccountUnavailable):
			status, message = 401, "บัญชีไม่พร้อมใช้งาน กรุณาเข้าสู่ระบบใหม่"
		case errors.Is(err, repositories.ErrSubscriptionRequired):
			status, message = 403, "กรุณามีแพ็กเกจที่เปิดใช้งานก่อนเพิ่มร้านค้า"
		case errors.Is(err, repositories.ErrShopLimit):
			status, message = 409, "จำนวนร้านค้าถึงโควตาของแพ็กเกจแล้ว"
		default:
			log.Printf("Create tenant shop: %v", err)
		}
		c.JSON(status, gin.H{"success": false, "message": message})
		return
	}
	c.JSON(201, gin.H{"success": true, "data": shop})
}
