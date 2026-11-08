package routes

import (
	"booking-saas/config"
	"booking-saas/controllers/tenant/bookings/ctrls"
	"booking-saas/controllers/tenant/bookings/repositories"
	"booking-saas/controllers/tenant/bookings/usecases"
	"booking-saas/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tenantBookingRoutes(tenant *gin.RouterGroup, db *pgxpool.Pool, cfg *config.Config) {
	ctrl := ctrls.NewBookingController(usecases.NewBookingUsecase(repositories.NewBookingRepository(db)))
	routes := tenant.Group("/shops/:shopID/bookings", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"))
	routes.GET("", ctrl.List)
	routes.POST("", ctrl.Create)
	routes.PATCH("/:bookingID", ctrl.Update)
	routes.PATCH("/:bookingID/status", ctrl.SetStatus)
	routes.PATCH("/:bookingID/customer", ctrl.UpdateCustomer)
}
