package routes

import (
	"booking-saas/controllers/tenant/services/ctrls"
	"booking-saas/controllers/tenant/services/repositories"
	"booking-saas/controllers/tenant/services/usecases"
	"booking-saas/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tenantServiceRoutes(active *gin.RouterGroup, db *pgxpool.Pool) {
	serviceRepository := repositories.NewServiceRepository(db)
	serviceUsecase := usecases.NewServiceUsecase(serviceRepository)
	serviceController := ctrls.NewServiceController(serviceUsecase)

	services := active.Group("/shops/:shopID/services", middleware.RequireShopAccess(db, true))
	services.GET("", serviceController.List)
	services.POST("", serviceController.Create)
	services.PATCH("/:serviceID", serviceController.Update)
}
