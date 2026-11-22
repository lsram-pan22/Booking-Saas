package routes

import (
	"time"

	"booking-saas/config"
	authCtrl "booking-saas/controllers/super_admin/auth/ctrls"
	authRepo "booking-saas/controllers/super_admin/auth/repositories"
	authUsecase "booking-saas/controllers/super_admin/auth/usecases"
	packageCtrl "booking-saas/controllers/super_admin/packages/ctrls"
	packageRepo "booking-saas/controllers/super_admin/packages/repositories"
	packageUsecase "booking-saas/controllers/super_admin/packages/usecases"
	accountCtrl "booking-saas/controllers/tenant/account/ctrls"
	accountRepo "booking-saas/controllers/tenant/account/repositories"
	accountUsecase "booking-saas/controllers/tenant/account/usecases"
	"booking-saas/middleware"
	"booking-saas/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SuperAdminRoutes(
	router *gin.RouterGroup,
	db *pgxpool.Pool,
	cfg *config.Config,
) {
	superAdmin := router.Group("/super-admin")
	requireJWT := middleware.RequireJWT(cfg.JWT.Secret)
	requireSuperAdmin := middleware.RequireRole(utils.RoleSuperAdmin)

	// Login and logout
	loginRepository := authRepo.NewUserRepository(db)
	loginUsecase := authUsecase.NewLoginUsecase(loginRepository, cfg)
	loginController := authCtrl.NewLoginController(loginUsecase)

	superAdmin.POST("/login", loginController.Login)
	superAdmin.POST("/logout",
		requireJWT,
		requireSuperAdmin,
		middleware.RequireSession(db, true, true),
		accountCtrl.LogoutAllSessions(db),
	)

	admin := superAdmin.Group("").Use(
		requireJWT,
		requireSuperAdmin,
		middleware.RequireSession(db, false),
	)

	// User management
	passwordRepository := accountRepo.NewPasswordRepository(db)
	passwordUsecase := accountUsecase.NewPasswordUsecase(passwordRepository)
	passwordController := accountCtrl.NewPasswordController(passwordUsecase)
	passwordLimit := middleware.NewRateLimiter(10, 15*time.Minute)

	admin.GET("/users", passwordController.Owners)
	admin.POST("/users/:userID/password/reset", passwordLimit.ByActor(), passwordController.Reset)

	// Package management
	packageRepository := packageRepo.NewPackageRepository(db)
	packages := packageUsecase.NewPackageUsecase(packageRepository)
	packageController := packageCtrl.NewPackageController(packages)

	admin.GET("/packages", packageController.List)
	admin.POST("/packages", packageController.Create)
	admin.PUT("/packages/:id", packageController.Update)
}
