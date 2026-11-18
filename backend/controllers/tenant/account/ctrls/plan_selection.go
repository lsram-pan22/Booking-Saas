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
	"time"
)

type PlanController struct{ Usecase *usecases.PlanUsecase }

func NewPlanController(u *usecases.PlanUsecase) *PlanController { return &PlanController{Usecase: u} }
func planError(c *gin.Context, err error) {
	if errors.Is(err, models.ErrPlanSelection) {
		c.JSON(403, gin.H{"success": false, "message": "เลือกร้านและพนักงานที่ใช้งานได้ตามโควตา Free โดยจำนวนผู้ใช้รวมเจ้าของร้านด้วย"})
		return
	}
	log.Printf("Subscription selection: %v", err)
	c.JSON(500, gin.H{"success": false, "message": "ไม่สามารถจัดการสิทธิ์แพ็กเกจได้"})
}
func (ctrl *PlanController) Overview(c *gin.Context) {
	claims, _ := middleware.AuthClaims(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	data, err := ctrl.Usecase.Repo.Overview(ctx, claims.UserID)
	if err != nil {
		planError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
func (ctrl *PlanController) Select(c *gin.Context) {
	claims, _ := middleware.AuthClaims(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1048576)
	var input models.PlanSelection
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(400, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := ctrl.Usecase.Select(ctx, claims.UserID, input); err != nil {
		planError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true})
}
