package routes

import (
	"booking-saas/config"
	superAdmin "booking-saas/controllers/super_admin/auth/ctrls"
	superAdminRepo "booking-saas/controllers/super_admin/auth/repositories"
	superAdminUsecase "booking-saas/controllers/super_admin/auth/usecases"
	packageCtrl "booking-saas/controllers/super_admin/packages/ctrls"
	packageRepo "booking-saas/controllers/super_admin/packages/repositories"
	packageUsecase "booking-saas/controllers/super_admin/packages/usecases"
	accountCtrl "booking-saas/controllers/tenant/account/ctrls"
	accountRepo "booking-saas/controllers/tenant/account/repositories"
	accountUsecase "booking-saas/controllers/tenant/account/usecases"
	"booking-saas/middleware"
	"booking-saas/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SuperAdminRoutes(
	router *gin.RouterGroup,
	db *pgxpool.Pool,
	cfg *config.Config,
) {
	auth := router.Group("/super-admin")
	passwordLimit := middleware.NewRateLimiter(10, 15*time.Minute)
	auth.POST("/logout", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole(utils.RoleSuperAdmin), middleware.RequireSession(db, true, true), accountCtrl.LogoutAllSessions(db))

	// Super Admin
	superAdminRepository := superAdminRepo.NewUserRepository(db)
	superAdminLogin := superAdminUsecase.NewLoginUsecase(
		superAdminRepository,
		cfg,
	)
	superAdminController := superAdmin.NewLoginController(superAdminLogin)

	auth.POST("/login", superAdminController.Login)

	// Package management requires a verified Super Admin token.
	admin := auth.Group("", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole(utils.RoleSuperAdmin), middleware.RequireSession(db, false))
	passwordCtrl := accountCtrl.NewPasswordController(accountUsecase.NewPasswordUsecase(accountRepo.NewPasswordRepository(db)))
	admin.GET("/users", passwordCtrl.Owners)
	admin.POST("/users/:userID/password/reset", passwordLimit.ByActor(), passwordCtrl.Reset)
	packageRepository := packageRepo.NewPackageRepository(db)
	packageService := packageUsecase.NewPackageUsecase(packageRepository)
	packageController := packageCtrl.NewPackageController(packageService)
	admin.GET("/packages", packageController.List)
	admin.POST("/packages", packageController.Create)
	admin.PUT("/packages/:id", packageController.Update)
}
