package ctrls

import (
	"errors"
	"log"
	"net/http"

	"booking-saas/controllers/auth/super_admin/models"
	"booking-saas/controllers/auth/super_admin/usecases"

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
	var req models.SuperAdminLoginRequest

	// รับและตรวจสอบ Request
	if err := c.ShouldBindJSON(&req); err != nil {
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

	// Login สำเร็จ
	c.JSON(http.StatusOK, models.SuperAdminLoginResponse{
		Success: true,
		Message: "Login successful",
		Data:    *data,
	})
}
