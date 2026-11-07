package routes

import (
	"booking-saas/config"
	"booking-saas/controllers/tenant/services/ctrls"
	"booking-saas/controllers/tenant/services/repositories"
	"booking-saas/controllers/tenant/services/usecases"
	"booking-saas/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tenantServiceRoutes(tenant *gin.RouterGroup, db *pgxpool.Pool, cfg *config.Config) {
	repo := repositories.NewServiceRepository(db)
	service := usecases.NewServiceUsecase(repo)
	ctrl := ctrls.NewServiceController(service)
	routes := tenant.Group("/shops/:shopID/services", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"))
	routes.GET("", ctrl.List)
	routes.POST("", ctrl.Create)
	routes.PATCH("/:serviceID", ctrl.Update)
}
