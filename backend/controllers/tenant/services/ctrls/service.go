package ctrls

import (
	"booking-saas/controllers/tenant/services/models"
	"booking-saas/controllers/tenant/services/repositories"
	"booking-saas/controllers/tenant/services/usecases"
	"booking-saas/middleware"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type ServiceController struct{ Usecase *usecases.ServiceUsecase }

func NewServiceController(u *usecases.ServiceUsecase) *ServiceController {
	return &ServiceController{Usecase: u}
}
func (ctrl *ServiceController) List(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	data, err := ctrl.Usecase.List(c.Request.Context(), claims.UserID, c.Param("shopID"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
func (ctrl *ServiceController) Create(c *gin.Context) { ctrl.save(c, "") }
func (ctrl *ServiceController) Update(c *gin.Context) { ctrl.save(c, c.Param("serviceID")) }
func (ctrl *ServiceController) save(c *gin.Context, serviceID string) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.ServiceRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"success": false, "message": "กรุณากรอกข้อมูลบริการให้ครบถ้วน"})
		return
	}
	data, err := ctrl.Usecase.Save(c.Request.Context(), claims.UserID, c.Param("shopID"), serviceID, req)
	if err != nil {
		respondError(c, err)
		return
	}
	status := http.StatusCreated
	if serviceID != "" {
		status = http.StatusOK
	}
	c.JSON(status, gin.H{"success": true, "data": data})
}
func respondError(c *gin.Context, err error) {
	status, message := 500, "ไม่สามารถดำเนินการกับบริการได้ในขณะนี้"
	switch {
	case errors.Is(err, usecases.ErrInvalidService):
		status, message = 400, "ข้อมูลบริการไม่ถูกต้อง กรุณาตรวจสอบชื่อ ระยะเวลา ราคา และสถานะ"
	case errors.Is(err, repositories.ErrNotFound):
		status, message = 404, "ไม่พบบริการ ร้านค้า หรือคุณไม่มีสิทธิ์ดำเนินการ"
	default:
		log.Printf("Shop services: %v", err)
	}
	c.JSON(status, gin.H{"success": false, "message": message})
}
