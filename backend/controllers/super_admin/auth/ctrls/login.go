package ctrls

import (
	"booking-saas/middleware"
	"errors"
	"log"
	"net/http"

	"booking-saas/controllers/super_admin/auth/models"
	"booking-saas/controllers/super_admin/auth/usecases"
	"booking-saas/utils"

	"github.com/gin-gonic/gin"
)

type LoginController struct {
	LoginUsecase *usecases.LoginUsecase
}

func NewLoginController(
	loginUsecase *usecases.LoginUsecase,
) *LoginController {
	return &LoginController{
		LoginUsecase: loginUsecase,
	}
}

func (ctrl *LoginController) Login(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
	var req models.SuperAdminLoginRequest

	// รับและตรวจสอบ Request
	if err := middleware.DecodeJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, models.SuperAdminLoginResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	// เรียก Usecase
	data, err := ctrl.LoginUsecase.Login(
		c.Request.Context(),
		req,
	)

	if err != nil {
		if errors.Is(err, usecases.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, models.SuperAdminLoginResponse{
				Success: false,
				Message: "Invalid email or password",
			})
			return
		}

		log.Printf("Super Admin login error: %v", err)

		c.JSON(http.StatusInternalServerError, models.SuperAdminLoginResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	c.Set(middleware.ClaimsKey, &utils.JWTClaims{UserID: data.User.ID, Role: data.User.Role})
	// Login สำเร็จ
	c.JSON(http.StatusOK, models.SuperAdminLoginResponse{
		Success: true,
		Message: "Login successful",
		Data:    *data,
	})
}
