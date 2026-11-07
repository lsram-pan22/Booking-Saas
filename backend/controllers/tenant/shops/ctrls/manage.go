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

type ManageController struct{ Usecase *usecases.ManageUsecase }

func NewManageController(u *usecases.ManageUsecase) *ManageController {
	return &ManageController{Usecase: u}
}
func (ctrl *ManageController) Get(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	shop, err := ctrl.Usecase.Get(c.Request.Context(), claims.UserID, c.Param("shopID"))
	if err != nil {
		manageError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": shop})
}
func (ctrl *ManageController) Update(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.UpdateShopRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"success": false, "message": "กรุณากรอกข้อมูลร้านค้าให้ถูกต้อง"})
		return
	}
	shop, err := ctrl.Usecase.Update(c.Request.Context(), claims.UserID, c.Param("shopID"), req)
	if err != nil {
		manageError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": shop})
}
func manageError(c *gin.Context, err error) {
	status, message := 500, "ไม่สามารถดำเนินการกับร้านค้าได้ในขณะนี้"
	switch {
	case errors.Is(err, usecases.ErrInvalidShop):
		status, message = 400, "ข้อมูลหรือรหัสร้านค้าไม่ถูกต้อง"
	case errors.Is(err, repositories.ErrShopNotFound):
		status, message = 404, "ไม่พบร้านค้าหรือคุณไม่มีสิทธิ์เข้าถึง"
	case errors.Is(err, usecases.ErrShopReadOnly):
		status, message = 403, "คุณไม่มีสิทธิ์แก้ไขร้านค้า หรือร้านถูกระงับหรือรอเปิดใช้งาน"
	default:
		log.Printf("Manage shop: %v", err)
	}
	c.JSON(status, gin.H{"success": false, "message": message})
}
