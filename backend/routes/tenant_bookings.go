package routes

import (
	"booking-saas/controllers/tenant/bookings/ctrls"
	"booking-saas/controllers/tenant/bookings/repositories"
	"booking-saas/controllers/tenant/bookings/usecases"
	"booking-saas/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tenantBookingRoutes(active *gin.RouterGroup, db *pgxpool.Pool) {
	bookingRepository := repositories.NewBookingRepository(db)
	bookingUsecase := usecases.NewBookingUsecase(bookingRepository)
	bookingController := ctrls.NewBookingController(bookingUsecase)

	bookings := active.Group("/shops/:shopID/bookings", middleware.RequireShopAccess(db, false))
	bookings.GET("", bookingController.List)
	bookings.GET("/:bookingID/edit-history", bookingController.EditHistory)
	bookings.POST("", bookingController.Create)
	bookings.PATCH("/:bookingID", bookingController.Update)
	bookings.PATCH("/:bookingID/status", bookingController.SetStatus)
	bookings.PATCH("/:bookingID/customer", bookingController.UpdateCustomer)
}
