package ctrls

import (
	"booking-saas/controllers/tenant/bookings/models"
	"booking-saas/controllers/tenant/bookings/repositories"
	"booking-saas/controllers/tenant/bookings/usecases"
	"booking-saas/middleware"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type BookingController struct{ Usecase *usecases.BookingUsecase }

func NewBookingController(u *usecases.BookingUsecase) *BookingController {
	return &BookingController{Usecase: u}
}
func (ctrl *BookingController) List(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	var data models.BookingList
	var err error
	switch c.Query("days") {
	case "7":
		data, err = ctrl.Usecase.ListWeek(c.Request.Context(), claims.UserID, c.Param("shopID"), c.Query("date"))
	case "", "1":
		data, err = ctrl.Usecase.List(c.Request.Context(), claims.UserID, c.Param("shopID"), c.Query("date"))
	default:
		err = usecases.ErrInvalid
	}
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
func (ctrl *BookingController) Create(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.CreateBookingRequest
	if c.ShouldBindJSON(&req) != nil {
		respondError(c, usecases.ErrInvalid)
		return
	}
	data, err := ctrl.Usecase.Create(c.Request.Context(), claims.UserID, c.Param("shopID"), req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(201, gin.H{"success": true, "data": data})
}
func (ctrl *BookingController) SetStatus(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.StatusRequest
	if c.ShouldBindJSON(&req) != nil {
		respondError(c, usecases.ErrInvalid)
		return
	}
	data, err := ctrl.Usecase.SetStatus(c.Request.Context(), claims.UserID, c.Param("shopID"), c.Param("bookingID"), req.Status, req.Remark)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
func respondError(c *gin.Context, err error) {
	status, message := 500, "ไม่สามารถดำเนินการกับการจองได้ในขณะนี้"
	switch {
	case errors.Is(err, repositories.ErrInvalidSchedule):
		status, message = 400, "วันและเวลาใหม่ต้องอยู่ในอนาคตไม่เกิน 2 ปี"
	case errors.Is(err, usecases.ErrInvalidCustomer):
		status, message = 400, "กรุณากรอกชื่อลูกค้าไม่เกิน 150 ตัวอักษร เบอร์โทรศัพท์ให้ถูกต้อง และหมายเหตุไม่เกิน 2,000 ตัวอักษร"
	case errors.Is(err, repositories.ErrReadOnly):
		status, message = 409, "แก้ไขได้เฉพาะการจองที่รอยืนยันหรือยืนยันแล้ว"
	case errors.Is(err, repositories.ErrOutsideHours):
		status, message = 409, "เวลาเริ่มและสิ้นสุดการจองต้องอยู่ในเวลาเปิด–ปิดร้าน"
	case errors.Is(err, usecases.ErrInvalidRemark):
		status, message = 400, "กรุณาระบุเหตุผลการยกเลิกไม่เกิน 1,000 ตัวอักษร"
	case errors.Is(err, usecases.ErrInvalid):
		status, message = 400, "ข้อมูลการจองไม่ถูกต้อง กรุณาตรวจสอบลูกค้า บริการ ผู้ให้บริการ และเวลาในอนาคตไม่เกิน 2 ปี"
	case errors.Is(err, repositories.ErrNotFound):
		status, message = 404, "ไม่พบการจอง ร้านค้า หรือคุณไม่มีสิทธิ์ดำเนินการ"
	case errors.Is(err, repositories.ErrUnavailable):
		status, message = 409, "บริการหรือผู้ให้บริการไม่พร้อมใช้งาน กรุณาโหลดข้อมูลใหม่"
	case errors.Is(err, repositories.ErrConflict):
		status, message = 409, "ผู้ให้บริการมีการจองในช่วงเวลานี้แล้ว กรุณาเลือกเวลาอื่น"
	case errors.Is(err, repositories.ErrTransition):
		status, message = 409, "สถานะการจองเปลี่ยนไปแล้ว หรือไม่สามารถเปลี่ยนเป็นสถานะที่เลือกได้"
	default:
		log.Printf("Shop bookings: %v", err)
	}
	c.JSON(status, gin.H{"success": false, "message": message})
}
