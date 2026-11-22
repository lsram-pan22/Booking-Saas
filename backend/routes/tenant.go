package routes

import (
	"time"

	"booking-saas/config"
	accountCtrl "booking-saas/controllers/tenant/account/ctrls"
	accountRepo "booking-saas/controllers/tenant/account/repositories"
	accountUsecase "booking-saas/controllers/tenant/account/usecases"
	shopCtrl "booking-saas/controllers/tenant/shops/ctrls"
	shopRepo "booking-saas/controllers/tenant/shops/repositories"
	shopUsecase "booking-saas/controllers/tenant/shops/usecases"
	"booking-saas/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TenantRoutes(
	router *gin.RouterGroup,
	db *pgxpool.Pool,
	cfg *config.Config,
) {
	tenant := router.Group("/tenant")
	requireJWT := middleware.RequireJWT(cfg.JWT.Secret)
	requireTenant := middleware.RequireRole("user")

	// Registration and authentication
	registerRepository := accountRepo.NewRegisterRepository(db)
	registerUsecase := accountUsecase.NewRegisterUsecase(registerRepository)
	registerController := accountCtrl.NewRegisterController(registerUsecase)

	loginRepository := accountRepo.NewLoginRepository(db)
	loginUsecase := accountUsecase.NewLoginUsecase(loginRepository, cfg)
	loginController := accountCtrl.NewLoginController(loginUsecase)

	tenant.POST("/register", registerController.Register)
	tenant.POST("/login", loginController.Login)
	tenant.POST("/logout",
		requireJWT,
		requireTenant,
		middleware.RequireSession(db, true, true),
		accountCtrl.LogoutAllSessions(db),
	)

	authenticated := tenant.Group("", requireJWT, requireTenant)
	account := authenticated.Group("", middleware.RequireSession(db, true))
	active := authenticated.Group("", middleware.RequireSession(db, false))

	// Account and passwords
	settingsRepository := accountRepo.NewSettingsRepository(db)
	settingsUsecase := accountUsecase.NewSettingsUsecase(settingsRepository)
	settingsController := accountCtrl.NewSettingsController(settingsUsecase)

	passwordRepository := accountRepo.NewPasswordRepository(db)
	passwordUsecase := accountUsecase.NewPasswordUsecase(passwordRepository)
	passwordController := accountCtrl.NewPasswordController(passwordUsecase)
	passwordLimit := middleware.NewRateLimiter(10, 15*time.Minute)

	account.GET("/me", loginController.Me)
	account.POST("/password/change", passwordLimit.ByActor(), passwordController.Change)
	active.PATCH("/me", settingsController.Update)

	// Subscription
	planRepository := accountRepo.NewPlanRepository(db)
	planUsecase := accountUsecase.NewPlanUsecase(planRepository)
	planController := accountCtrl.NewPlanController(planUsecase)

	active.GET("/subscription", planController.Overview)
	active.PUT("/subscription/selection", planController.Select)

	// Shops and members
	shopRepository := shopRepo.NewShopRepository(db)
	shopService := shopUsecase.NewShopUsecase(shopRepository)
	shopController := shopCtrl.NewShopController(shopService)

	manageUsecase := shopUsecase.NewManageUsecase(shopRepository)
	manageController := shopCtrl.NewManageController(manageUsecase)
	membersUsecase := shopUsecase.NewMembersUsecase(shopRepository)
	membersController := shopCtrl.NewMembersController(membersUsecase)
	addMemberUsecase := shopUsecase.NewAddMemberUsecase(shopRepository)
	addMemberController := shopCtrl.NewAddMemberController(addMemberUsecase)
	memberStatusUsecase := shopUsecase.NewMemberStatusUsecase(shopRepository)
	memberStatusController := shopCtrl.NewMemberStatusController(memberStatusUsecase)

	active.POST("/shops", shopController.Create)

	shop := active.Group("/shops/:shopID", middleware.RequireShopAccess(db, true))
	shop.GET("", manageController.Get)
	shop.PATCH("", manageController.Update)
	shop.GET("/members", membersController.List)
	shop.POST("/members", addMemberController.Add)
	shop.PATCH("/members/:memberID", memberStatusController.Set)
	shop.POST("/users/:userID/password/reset", passwordLimit.ByActor(), passwordController.Reset)

	tenantServiceRoutes(active, db)
	tenantBookingRoutes(active, db)
}
