package routes

import (
	"booking-saas/config"
	superAdmin "booking-saas/controllers/auth/super_admin/ctrls"
	superAdminRepo "booking-saas/controllers/auth/super_admin/repositories"
	superAdminUsecase "booking-saas/controllers/auth/super_admin/usecases"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRoutes(
	router *gin.RouterGroup,
	db *pgxpool.Pool,
	cfg *config.Config,
) {
	auth := router.Group("/auth")

	// Super Admin
	superAdminRepository := superAdminRepo.NewUserRepository(db)
	superAdminLogin := superAdminUsecase.NewLoginUsecase(
		superAdminRepository,
		cfg,
	)
	superAdminController := superAdmin.NewLoginController(superAdminLogin)

	auth.POST("/super-admin/login", superAdminController.Login)
}
