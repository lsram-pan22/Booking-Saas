package ctrls

import (
	"booking-saas/controllers/tenant/bookings/models"
	"booking-saas/controllers/tenant/bookings/usecases"
	"booking-saas/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (ctrl *BookingController) Update(c *gin.Context) {
	claims, ok := middleware.AuthClaims(c)
	if !ok {
		c.AbortWithStatus(401)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.UpdateBookingRequest
	if c.ShouldBindJSON(&req) != nil {
		respondError(c, usecases.ErrInvalid)
		return
	}
	data, err := ctrl.Usecase.Update(c.Request.Context(), claims.UserID, c.Param("shopID"), c.Param("bookingID"), req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
