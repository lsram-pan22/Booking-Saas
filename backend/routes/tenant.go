package routes

import (
	"booking-saas/config"
	"booking-saas/controllers/tenant/account/ctrls"
	"booking-saas/controllers/tenant/account/repositories"
	"booking-saas/controllers/tenant/account/usecases"
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
	auth := router.Group("/tenant")

	repo := repositories.NewRegisterRepository(db)
	usecase := usecases.NewRegisterUsecase(repo)
	ctrl := ctrls.NewRegisterController(usecase)

	auth.POST("/register", ctrl.Register)
	loginRepo := repositories.NewLoginRepository(db)
	loginUsecase := usecases.NewLoginUsecase(loginRepo, cfg)
	loginCtrl := ctrls.NewLoginController(loginUsecase)
	auth.POST("/login", loginCtrl.Login)
	auth.GET("/me", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"), loginCtrl.Me)
	settingsRepo := repositories.NewSettingsRepository(db)
	settingsUsecase := usecases.NewSettingsUsecase(settingsRepo)
	settingsCtrl := ctrls.NewSettingsController(settingsUsecase)
	auth.PATCH("/me", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"), settingsCtrl.Update)
	shopsRepository := shopRepo.NewShopRepository(db)
	shopsUsecase := shopUsecase.NewShopUsecase(shopsRepository)
	shopsController := shopCtrl.NewShopController(shopsUsecase)
	auth.POST("/shops", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"), shopsController.Create)
	manageUsecase := shopUsecase.NewManageUsecase(shopsRepository)
	manageController := shopCtrl.NewManageController(manageUsecase)
	auth.GET("/shops/:shopID", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"), manageController.Get)
	auth.PATCH("/shops/:shopID", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"), manageController.Update)
	membersUsecase := shopUsecase.NewMembersUsecase(shopsRepository)
	membersController := shopCtrl.NewMembersController(membersUsecase)
	auth.GET("/shops/:shopID/members", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"), membersController.List)
	addMemberUsecase := shopUsecase.NewAddMemberUsecase(shopsRepository)
	addMemberController := shopCtrl.NewAddMemberController(addMemberUsecase)
	auth.POST("/shops/:shopID/members", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"), addMemberController.Add)
	memberStatusUsecase := shopUsecase.NewMemberStatusUsecase(shopsRepository)
	memberStatusController := shopCtrl.NewMemberStatusController(memberStatusUsecase)
	auth.PATCH("/shops/:shopID/members/:memberID", middleware.RequireJWT(cfg.JWT.Secret), middleware.RequireRole("user"), memberStatusController.Set)
}
